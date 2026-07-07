package reader

import (
	"slices"
	"strings"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/meta"
	"github.com/bensadeh/circumflex/scrollbar"
	"github.com/bensadeh/circumflex/view/message"
	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

const sectionMarker = "■"

type Meta struct {
	URL       string
	Author    string
	TimeAgo   string
	ID        int
	Points    int
	NerdFonts bool
}

type Model struct {
	pane.Scroller

	keymap keyMap

	headerLines []int // line indices containing ■ (section headers)
	title       string
	titleHeader string
	screenWidth int
	showHelp    bool

	parsed      *article.Parsed // nil when created with pre-rendered content
	maxWidth    int
	articleMeta Meta
}

const (
	headerHeight = 2 // title + overline separator
	footerHeight = 2 // underline separator + blank line, mirroring the comment view's footer
)

func newFromContent(content, title string, width, height int, articleMeta Meta) *Model {
	m := &Model{
		keymap:      defaultKeyMap(),
		title:       title,
		screenWidth: width,
		articleMeta: articleMeta,
	}

	m.initViewport(content, width, height)

	return m
}

// NewWithArticle creates a reader view that can re-render on resize.
func NewWithArticle(parsed *article.Parsed, title string, maxWidth int, width, height int, articleMeta Meta) *Model {
	contentWidth := layout.ReaderContentWidth(width, maxWidth)
	header := meta.ReaderModeMetaBlock(articleMeta.URL, articleMeta.Author, articleMeta.TimeAgo, articleMeta.ID, articleMeta.Points, articleMeta.NerdFonts, contentWidth)
	content := parsed.RenderWithHeader(contentWidth, width, header)

	m := &Model{
		keymap:      defaultKeyMap(),
		title:       title,
		screenWidth: width,
		parsed:      parsed,
		maxWidth:    maxWidth,
		articleMeta: articleMeta,
	}

	m.initViewport(content, width, height)

	return m
}

// DisableStoryNavigation removes the J/K adjacent-story bindings, for
// standalone use where there is no story list to move through.
func (m *Model) DisableStoryNavigation() {
	m.keymap.DisableStoryNavigation()
}

func (m *Model) initViewport(content string, width, height int) {
	m.Viewport = pane.NewViewport(width, height-headerHeight-footerHeight)
	m.setContent(content)
	m.rebuildTitleHeader()
}

func (m *Model) setContent(content string) {
	trimmed := strings.TrimRight(content, "\n")
	lines := strings.Split(trimmed, "\n")

	m.ContentLines = len(lines)
	m.headerLines = nil

	for i, line := range lines {
		if strings.Contains(line, sectionMarker) {
			m.headerLines = append(m.headerLines, i)
		}
	}

	// Add bottom padding so G scrolls the last content line to the bottom.
	m.Viewport.SetContent(trimmed + strings.Repeat("\n", m.Viewport.Height()))
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)

	case tea.MouseWheelMsg:
		if m.showHelp {
			return nil
		}

		m.HandleMouseWheel(msg)

		return nil

	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.Viewport.SetWidth(msg.Width)
		m.Viewport.SetHeight(max(0, msg.Height-headerHeight-footerHeight))

		m.rebuildTitleHeader()

		if m.parsed != nil {
			m.rerender()
		}

		return nil
	}

	return m.Forward(msg)
}

// rerender re-renders the article at the current screen width and updates
// the viewport content, preserving the scroll position.
func (m *Model) rerender() {
	yOffset := m.Viewport.YOffset()

	contentWidth := layout.ReaderContentWidth(m.screenWidth, m.maxWidth)
	hdr := meta.ReaderModeMetaBlock(m.articleMeta.URL, m.articleMeta.Author, m.articleMeta.TimeAgo, m.articleMeta.ID, m.articleMeta.Points, m.articleMeta.NerdFonts, contentWidth)
	m.setContent(m.parsed.RenderWithHeader(contentWidth, m.screenWidth, hdr))

	maxOffset := max(0, m.ContentLines-m.Viewport.Height())
	m.Viewport.SetYOffset(min(yOffset, maxOffset))
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if m.showHelp {
		if key.Matches(msg, m.keymap.Quit, m.keymap.Help) {
			m.showHelp = false
		}

		return nil
	}

	switch {
	case key.Matches(msg, m.keymap.Quit):
		return func() tea.Msg { return message.ReaderViewQuit{} }

	case key.Matches(msg, m.keymap.GotoTop):
		m.Viewport.GotoTop()

	case key.Matches(msg, m.keymap.GotoBottom):
		m.GotoBottom()

	case key.Matches(msg, m.keymap.HalfPageDown):
		m.HalfPageDown()

	case key.Matches(msg, m.keymap.HalfPageUp):
		m.HalfPageUp()

	case key.Matches(msg, m.keymap.PageDown):
		m.PageDown()

	case key.Matches(msg, m.keymap.PageUp):
		m.PageUp()

	case key.Matches(msg, m.keymap.NextHeader):
		m.jumpToHeader(1)

	case key.Matches(msg, m.keymap.PrevHeader):
		m.jumpToHeader(-1)

	case key.Matches(msg, m.keymap.Help):
		m.showHelp = true

	case key.Matches(msg, m.keymap.OpenLink):
		return pane.OpenStoryInBrowser(m.articleMeta.URL, m.articleMeta.ID)

	case key.Matches(msg, m.keymap.OpenComments):
		return pane.OpenCommentsInBrowser(m.articleMeta.ID)

	case key.Matches(msg, m.keymap.NextStory):
		return message.OpenAdjacentStoryCmd(1)

	case key.Matches(msg, m.keymap.PrevStory):
		return message.OpenAdjacentStoryCmd(-1)

	default:
		return m.Forward(msg)
	}

	return nil
}

func (m *Model) View() string {
	if m.showHelp {
		content := help.FitToHeight(
			help.ReaderHelpScreen(m.screenWidth, m.keymap.NextStory.Enabled()),
			m.Viewport.Height(),
		)

		return header.HelpHeader("Reader Mode", m.screenWidth) + "\n" +
			content + "\n" +
			pane.FooterSeparator(m.screenWidth) + "\n" +
			help.Footer(m.screenWidth)
	}

	content := scrollbar.Attach(m.Viewport.View(), m.screenWidth, m.ContentLines, m.Viewport.Height(), m.Viewport.YOffset())

	return m.titleHeader + "\n" + content + "\n" + pane.FooterSeparator(m.screenWidth) + "\n"
}

func (m *Model) rebuildTitleHeader() {
	m.titleHeader = pane.TitleHeader(m.title, m.articleMeta.NerdFonts, layout.ReaderViewLeftMargin, m.screenWidth)
}

func (m *Model) jumpToHeader(direction int) {
	if len(m.headerLines) == 0 {
		return
	}

	yOffset := m.Viewport.YOffset()

	if direction > 0 {
		for _, line := range m.headerLines {
			if line > yOffset {
				m.Viewport.SetYOffset(line)

				return
			}
		}
	} else {
		for _, v := range slices.Backward(m.headerLines) {
			if v < yOffset {
				m.Viewport.SetYOffset(v)

				return
			}
		}

		if yOffset > 0 {
			m.Viewport.SetYOffset(0)
		}
	}
}

func Run(content, title string, articleMeta Meta) error {
	return pane.RunStandalone(func(width, height int) pane.View {
		m := newFromContent(content, title, width, height, articleMeta)
		m.DisableStoryNavigation()

		return m
	})
}
