package reader

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/meta"
	"github.com/bensadeh/circumflex/scrollbar"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/syntax"
	"github.com/bensadeh/circumflex/view/message"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
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
	viewport viewport.Model
	keymap   keyMap

	headerLines    []int // line indices containing ■ (section headers)
	title          string
	titleHeader    string
	contentLines   int // excludes bottom padding
	screenWidth    int
	viewportHeight int
	standalone     bool // quit sends tea.Quit instead of ReaderViewQuit
	showHelp       bool

	parsed      *article.Parsed // nil when created with pre-rendered content
	maxWidth    int
	articleMeta Meta
}

const (
	headerHeight = 2 // title + overline separator
	footerHeight = 2 // underline separator + blank line, mirroring the comment view's footer
)

func newFromContent(content, title string, width, height int) *Model {
	m := &Model{
		keymap:      defaultKeyMap(),
		title:       title,
		screenWidth: width,
	}

	m.initViewport(content, width, height)

	return m
}

// NewWithArticle creates a reader view that can re-render on resize.
func NewWithArticle(parsed *article.Parsed, title string, maxWidth int, width, height int, articleMeta Meta) *Model {
	contentWidth := layout.ReaderContentWidth(width, maxWidth)
	header := meta.ReaderModeMetaBlock(articleMeta.URL, articleMeta.Author, articleMeta.TimeAgo, articleMeta.ID, articleMeta.Points, articleMeta.NerdFonts, contentWidth)
	content := parsed.RenderWithHeader(contentWidth, header)

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
	m.keymap.NextStory.SetEnabled(false)
	m.keymap.PrevStory.SetEnabled(false)
}

func (m *Model) initViewport(content string, width, height int) {
	vpHeight := max(0, height-headerHeight-footerHeight)

	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(vpHeight),
	)

	vp.KeyMap = viewport.DefaultKeyMap()
	vp.KeyMap.HalfPageDown.SetEnabled(false)
	vp.KeyMap.HalfPageUp.SetEnabled(false)
	vp.KeyMap.PageDown.SetEnabled(false)
	vp.KeyMap.PageUp.SetEnabled(false)
	vp.MouseWheelEnabled = false

	m.viewport = vp
	m.viewportHeight = vpHeight
	m.setContent(content)
	m.rebuildTitleHeader()
}

func (m *Model) setContent(content string) {
	trimmed := strings.TrimRight(content, "\n")
	lines := strings.Split(trimmed, "\n")

	m.contentLines = len(lines)
	m.headerLines = nil

	for i, line := range lines {
		if strings.Contains(line, sectionMarker) {
			m.headerLines = append(m.headerLines, i)
		}
	}

	// Add bottom padding so G scrolls the last content line to the bottom.
	m.viewport.SetContent(trimmed + strings.Repeat("\n", m.viewportHeight))
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

		delta := m.viewport.MouseWheelDelta
		maxOffset := max(0, m.contentLines-m.viewportHeight)

		switch msg.Button {
		case tea.MouseWheelDown:
			m.viewport.SetYOffset(min(m.viewport.YOffset()+delta, maxOffset))
		case tea.MouseWheelUp:
			m.viewport.SetYOffset(max(0, m.viewport.YOffset()-delta))
		}

		return nil

	case tea.WindowSizeMsg:
		m.screenWidth = msg.Width
		m.viewportHeight = msg.Height - headerHeight - footerHeight
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(m.viewportHeight)

		m.rebuildTitleHeader()

		if m.parsed != nil {
			m.rerender()
		}

		return nil
	}

	before := m.viewport.YOffset()

	var cmd tea.Cmd

	m.viewport, cmd = m.viewport.Update(msg)
	m.clampScroll(before)

	return cmd
}

// rerender re-renders the article at the current screen width and updates
// the viewport content, preserving the scroll position.
func (m *Model) rerender() {
	yOffset := m.viewport.YOffset()

	contentWidth := layout.ReaderContentWidth(m.screenWidth, m.maxWidth)
	hdr := meta.ReaderModeMetaBlock(m.articleMeta.URL, m.articleMeta.Author, m.articleMeta.TimeAgo, m.articleMeta.ID, m.articleMeta.Points, m.articleMeta.NerdFonts, contentWidth)
	m.setContent(m.parsed.RenderWithHeader(contentWidth, hdr))

	maxOffset := max(0, m.contentLines-m.viewportHeight)
	m.viewport.SetYOffset(min(yOffset, maxOffset))
}

func (m *Model) handleKeyPress(msg tea.KeyPressMsg) tea.Cmd {
	if m.showHelp {
		if key.Matches(msg, m.keymap.Quit, m.keymap.Help) {
			m.showHelp = false
		}

		return nil
	}

	if key.Matches(msg, m.keymap.Quit) {
		if m.standalone {
			return tea.Quit
		}

		return func() tea.Msg { return message.ReaderViewQuit{} }
	}

	if key.Matches(msg, m.keymap.GotoTop) {
		m.viewport.GotoTop()

		return nil
	}

	if key.Matches(msg, m.keymap.GotoBottom) {
		m.gotoBottom()

		return nil
	}

	if key.Matches(msg, m.keymap.HalfPageDown) {
		m.halfPageDown()

		return nil
	}

	if key.Matches(msg, m.keymap.HalfPageUp) {
		m.halfPageUp()

		return nil
	}

	if key.Matches(msg, m.keymap.PageDown) {
		m.pageDown()

		return nil
	}

	if key.Matches(msg, m.keymap.PageUp) {
		m.pageUp()

		return nil
	}

	if key.Matches(msg, m.keymap.NextHeader) {
		m.jumpToHeader(1)

		return nil
	}

	if key.Matches(msg, m.keymap.PrevHeader) {
		m.jumpToHeader(-1)

		return nil
	}

	if key.Matches(msg, m.keymap.Help) {
		m.showHelp = true

		return nil
	}

	if key.Matches(msg, m.keymap.OpenLink) {
		return m.openStoryInBrowser()
	}

	if key.Matches(msg, m.keymap.OpenComments) {
		return m.openCommentsInBrowser()
	}

	if key.Matches(msg, m.keymap.NextStory) {
		return message.OpenAdjacentStoryCmd(1)
	}

	if key.Matches(msg, m.keymap.PrevStory) {
		return message.OpenAdjacentStoryCmd(-1)
	}

	before := m.viewport.YOffset()

	var cmd tea.Cmd

	m.viewport, cmd = m.viewport.Update(msg)
	m.clampScroll(before)

	return cmd
}

func (m *Model) View() string {
	if m.showHelp {
		content := help.FitToHeight(
			help.ReaderHelpScreen(m.screenWidth, m.keymap.NextStory.Enabled()),
			m.viewportHeight,
		)

		return header.HelpHeader("Reader Mode", m.screenWidth) + "\n" +
			content + "\n" +
			m.footerSeparator() + "\n" +
			help.Footer(m.screenWidth)
	}

	content := scrollbar.Attach(m.viewport.View(), m.screenWidth, m.contentLines, m.viewportHeight, m.viewport.YOffset())

	return m.titleHeader + "\n" + content + "\n" + m.footerSeparator() + "\n"
}

func (m *Model) rebuildTitleHeader() {
	leftMargin := strings.Repeat(" ", layout.ReaderViewLeftMargin)
	maxTitleWidth := max(0, m.screenWidth-layout.ReaderViewLeftMargin)
	title := syntax.ReplaceSpecialContentTags(m.title, m.articleMeta.NerdFonts)
	title = xansi.Truncate(title, maxTitleWidth, "…")

	title = syntax.HighlightYCStartupsInHeadlines(title, syntax.HeadlineInCommentSection, m.articleMeta.NerdFonts)
	title = syntax.HighlightYear(title, syntax.HeadlineInCommentSection)
	title = syntax.HighlightHackerNewsHeadlines(title, syntax.HeadlineInCommentSection)
	title = syntax.HighlightSpecialContent(title, syntax.HeadlineInCommentSection, m.articleMeta.NerdFonts)

	title = leftMargin + ansi.Bold + title + ansi.Reset
	separator := header.Underline(m.screenWidth)

	m.titleHeader = title + "\n" + separator
}

func (m *Model) footerSeparator() string {
	s := lipgloss.NewStyle().Underline(true).Width(m.screenWidth)
	if header.MemorialActive() {
		s = s.Foreground(style.MemorialColor())
	}

	return s.Render(strings.Repeat(" ", m.screenWidth))
}

// clampScroll prevents scrolling down past the last content line while still
// allowing upward scrolling from a position beyond the clamp point (e.g. after
// an n/N header jump).
func (m *Model) clampScroll(before int) {
	maxOffset := max(0, m.contentLines-m.viewportHeight)
	after := m.viewport.YOffset()

	if after > before && after > maxOffset {
		m.viewport.SetYOffset(max(before, maxOffset))
	}
}

func (m *Model) halfPageDown() {
	halfPage := m.viewportHeight / 2
	maxOffset := max(0, m.contentLines-m.viewportHeight)
	m.viewport.SetYOffset(min(m.viewport.YOffset()+halfPage, maxOffset))
}

func (m *Model) halfPageUp() {
	halfPage := m.viewportHeight / 2
	m.viewport.SetYOffset(max(0, m.viewport.YOffset()-halfPage))
}

func (m *Model) pageDown() {
	maxOffset := max(0, m.contentLines-m.viewportHeight)
	m.viewport.SetYOffset(min(m.viewport.YOffset()+m.viewportHeight, maxOffset))
}

func (m *Model) pageUp() {
	m.viewport.SetYOffset(max(0, m.viewport.YOffset()-m.viewportHeight))
}

func (m *Model) gotoBottom() {
	m.viewport.SetYOffset(max(0, m.contentLines-m.viewportHeight))
}

func (m *Model) openStoryInBrowser() tea.Cmd {
	url := m.articleMeta.URL
	if url == "" {
		if m.articleMeta.ID == 0 {
			return nil
		}

		url = hn.ItemURL(m.articleMeta.ID)
	}

	return message.OpenInBrowser(url)
}

func (m *Model) openCommentsInBrowser() tea.Cmd {
	if m.articleMeta.ID == 0 {
		return nil
	}

	return message.OpenInBrowser(hn.ItemURL(m.articleMeta.ID))
}

func (m *Model) jumpToHeader(direction int) {
	if len(m.headerLines) == 0 {
		return
	}

	yOffset := m.viewport.YOffset()

	if direction > 0 {
		for _, line := range m.headerLines {
			if line > yOffset {
				m.viewport.SetYOffset(line)

				return
			}
		}
	} else {
		for _, v := range slices.Backward(m.headerLines) {
			if v < yOffset {
				m.viewport.SetYOffset(v)

				return
			}
		}

		if yOffset > 0 {
			m.viewport.SetYOffset(0)
		}
	}
}

type standaloneModel struct {
	inner      *Model
	browserErr error
}

func (s standaloneModel) Init() tea.Cmd {
	return s.inner.Init()
}

func (s standaloneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if failed, ok := msg.(message.BrowserOpenFailed); ok {
		s.browserErr = failed.Err
	}

	cmd := s.inner.Update(msg)

	return s, cmd
}

func (s standaloneModel) View() tea.View {
	v := tea.NewView(s.inner.View())
	v.AltScreen = true

	return v
}

func Run(content, title string, articleMeta Meta) error {
	m := newFromContent(content, title, 0, 0)
	m.standalone = true
	m.articleMeta = articleMeta
	m.DisableStoryNavigation()

	p := tea.NewProgram(standaloneModel{inner: m})

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	if sm, ok := finalModel.(standaloneModel); ok && sm.browserErr != nil {
		fmt.Fprintf(os.Stderr, "circumflex: could not open browser: %v\n", sm.browserErr)
	}

	return nil
}
