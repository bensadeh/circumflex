package reader

import (
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/meta"
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

// Meta holds HN metadata displayed in the reader mode header block.
type Meta struct {
	URL       string
	Author    string
	TimeAgo   string
	ID        int
	Points    int
	NerdFonts bool
}

// Model is the Bubble Tea model for the built-in reader view.
type Model struct {
	viewport viewport.Model
	keymap   keyMap

	headerLines    []int  // line indices containing ■ (section headers)
	title          string // article title for the fixed header
	contentLines   int    // actual content lines (excluding bottom padding)
	screenWidth    int
	viewportHeight int
	standalone     bool // when true, quit sends tea.Quit instead of ReaderViewQuitMsg
	showHelp       bool

	// Fields for re-rendering on resize.
	parsed      *article.Parsed // nil when created with pre-rendered content
	maxWidth    int             // ArticleWidth cap
	articleMeta Meta
}

const (
	headerHeight = 2 // title + overline separator
	footerHeight = 2 // underline separator + keybinding hints
)

// New creates a new reader view model with pre-rendered content.
func New(content, title string, width, height int) *Model {
	m := &Model{
		keymap:      defaultKeyMap(),
		title:       title,
		screenWidth: width,
	}

	m.initViewport(content, width, height)

	return m
}

// NewWithArticle creates a reader view that can re-render on resize.
func NewWithArticle(parsed *article.Parsed, title string, maxWidth int, width, height int, m2 Meta) *Model {
	contentWidth := layout.ReaderContentWidth(width, maxWidth)
	header := meta.ReaderModeMetaBlock(m2.URL, m2.Author, m2.TimeAgo, m2.ID, m2.Points, m2.NerdFonts, contentWidth)
	content := parsed.RenderWithHeader(contentWidth, header)

	m := &Model{
		keymap:      defaultKeyMap(),
		title:       title,
		screenWidth: width,
		parsed:      parsed,
		maxWidth:    maxWidth,
		articleMeta: m2,
	}

	m.initViewport(content, width, height)

	return m
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

	trimmed := strings.TrimRight(content, "\n")
	lines := strings.Split(trimmed, "\n")
	contentLineCount := len(lines)

	// Scan for header lines (lines containing the ■ block character).
	var headers []int

	for i, line := range lines {
		if strings.Contains(line, sectionMarker) {
			headers = append(headers, i)
		}
	}

	// Add bottom padding so G scrolls the last content line to the bottom.
	padding := strings.Repeat("\n", vpHeight)
	padded := trimmed + padding

	vp.SetContent(padded)

	m.viewport = vp
	m.headerLines = headers
	m.contentLines = contentLineCount
	m.viewportHeight = vpHeight
}

// Init returns nil; no initial commands needed.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the reader view.
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
	content := m.parsed.RenderWithHeader(contentWidth, hdr)
	trimmed := strings.TrimRight(content, "\n")
	lines := strings.Split(trimmed, "\n")

	m.contentLines = len(lines)

	var headers []int

	for i, line := range lines {
		if strings.Contains(line, sectionMarker) {
			headers = append(headers, i)
		}
	}

	m.headerLines = headers

	padding := strings.Repeat("\n", m.viewportHeight)
	m.viewport.SetContent(trimmed + padding)

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

		return func() tea.Msg { return message.ReaderViewQuitMsg{} }
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

	before := m.viewport.YOffset()

	var cmd tea.Cmd

	m.viewport, cmd = m.viewport.Update(msg)
	m.clampScroll(before)

	return cmd
}

// View renders the reader view.
func (m *Model) View() string {
	if m.showHelp {
		content := help.FitToHeight(
			help.ReaderHelpScreen(m.screenWidth),
			m.viewportHeight,
		)

		return header.HelpHeader("Reader Mode", m.screenWidth) + "\n" +
			content + "\n" +
			m.footerSeparator() + "\n" +
			help.Footer(m.screenWidth)
	}

	return m.headerView() + "\n" + m.viewport.View() + "\n" + m.footerSeparator() + "\n" + m.modeIndicator()
}

func (m *Model) headerView() string {
	leftMargin := strings.Repeat(" ", layout.ReaderViewLeftMargin)
	maxTitleWidth := max(0, m.screenWidth-layout.ReaderViewLeftMargin)
	title := syntax.ReplaceSpecialContentTags(m.title, m.articleMeta.NerdFonts)
	title = xansi.Truncate(title, maxTitleWidth, "…")

	title = syntax.HighlightYCStartupsInHeadlines(title, syntax.HeadlineInCommentSection, m.articleMeta.NerdFonts)
	title = syntax.HighlightYear(title, syntax.HeadlineInCommentSection)
	title = syntax.HighlightHackerNewsHeadlines(title, syntax.HeadlineInCommentSection)
	title = syntax.HighlightSpecialContent(title, syntax.HeadlineInCommentSection, m.articleMeta.NerdFonts)

	title = leftMargin + ansi.Bold + title + ansi.Reset
	separator := strings.Repeat("‾", m.screenWidth)

	return title + "\n" + separator
}

func (m *Model) footerSeparator() string {
	return lipgloss.NewStyle().Underline(true).
		Width(m.screenWidth).
		Render(strings.Repeat(" ", m.screenWidth))
}

func (m *Model) modeIndicator() string {
	left := style.ModeIndicator(nil)
	right := style.RenderBinding(style.Binding{Key: "i", Desc: "help"})

	totalWidth := m.screenWidth
	if m.maxWidth > 0 {
		contentWidth := layout.ReaderContentWidth(m.screenWidth, m.maxWidth)
		totalWidth = layout.ReaderViewLeftMargin + contentWidth
	}

	padding := max(1, totalWidth-lipgloss.Width(left)-lipgloss.Width(right))

	return left + strings.Repeat(" ", padding) + right
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
		for i := len(m.headerLines) - 1; i >= 0; i-- {
			if m.headerLines[i] < yOffset {
				m.viewport.SetYOffset(m.headerLines[i])

				return
			}
		}

		if yOffset > 0 {
			m.viewport.SetYOffset(0)
		}
	}
}

// standaloneModel wraps Model to implement tea.Model for standalone use.
type standaloneModel struct {
	inner *Model
}

func (s standaloneModel) Init() tea.Cmd {
	return s.inner.Init()
}

func (s standaloneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := s.inner.Update(msg)

	return s, cmd
}

func (s standaloneModel) View() tea.View {
	v := tea.NewView(s.inner.View())
	v.AltScreen = true

	return v
}

// Run launches the reader as a standalone Bubble Tea program.
func Run(content, title string) error {
	m := New(content, title, 0, 0)
	m.standalone = true

	p := tea.NewProgram(standaloneModel{inner: m})
	_, err := p.Run()

	return err
}
