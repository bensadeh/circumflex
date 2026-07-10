package reader

import (
	"image/color"
	"slices"
	"strings"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/scrollbar"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/view/message"
	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

const sectionMarker = "■"

// Options carries the reader's display knobs and the story identity its
// browser-opening keys target. What the header above the article shows is
// not the reader's concern: callers inject that via buildHeader.
type Options struct {
	URL       string
	ID        int
	NerdFonts bool
	Images    bool
	TermBG    color.Color // terminal background when already known, for image transparency
}

type Model struct {
	pane.Scroller

	keymap keyMap

	headerLines []int // line indices containing ■ (section headers)
	title       string
	titleHeader string
	paneWidth   int
	showHelp    bool

	parsed      *article.Parsed // nil when created with pre-rendered content
	maxWidth    int
	opts        Options
	showImages  bool
	termBG      color.Color // nil until the terminal reports it
	buildHeader func(contentWidth int) string
	blockStarts []int // line index of each article block in the current render
}

// newFromContent builds a reader over pre-rendered content with no parsed
// article, used by tests to exercise the viewport and scrolling directly.
func newFromContent(content, title string, width, height int) *Model {
	m := &Model{
		keymap:    defaultKeyMap(),
		title:     title,
		paneWidth: width,
	}

	m.initViewport(content, width, height)

	return m
}

// NewWithArticle creates a reader view that can re-render on resize.
// buildHeader supplies the block drawn above the article for a given content
// width; nil renders the article with no header.
func NewWithArticle(parsed *article.Parsed, title string, maxWidth int, width, height int, opts Options, buildHeader func(contentWidth int) string) *Model {
	m := &Model{
		keymap:      defaultKeyMap(),
		title:       title,
		paneWidth:   width,
		parsed:      parsed,
		maxWidth:    maxWidth,
		opts:        opts,
		showImages:  opts.Images,
		termBG:      opts.TermBG,
		buildHeader: buildHeader,
	}

	content, starts := m.renderArticle()
	m.blockStarts = starts

	m.initViewport(content, width, height)

	return m
}

// renderArticle renders the article at the current pane width, prefixed with
// its header and a separating blank line, returning the content and the
// blocks' starting lines. The single source of the width derivation shared
// by the initial render and every resize.
func (m *Model) renderArticle() (string, []int) {
	contentWidth := layout.ReaderContentWidth(m.paneWidth, m.maxWidth)
	images := article.ImageOptions{Show: m.showImages, TerminalBG: m.termBG}

	header := ""

	if m.buildHeader != nil {
		margin := strings.Repeat(" ", layout.ReaderViewLeftMargin)
		header = style.PrefixLines(m.buildHeader(contentWidth), margin) + "\n\n"
	}

	return m.parsed.RenderWithHeader(contentWidth, m.paneWidth, header, images)
}

// DisableStoryNavigation removes the J/K adjacent-story bindings, for
// standalone use where there is no story list to move through.
func (m *Model) DisableStoryNavigation() {
	m.keymap.DisableStoryNavigation()
}

func (m *Model) initViewport(content string, width, height int) {
	m.Viewport = pane.NewViewport(width, height-layout.PaneChromeHeight)
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
	// Handing over the existing split spares the viewport from re-splitting
	// megabytes of image art.
	m.Viewport.SetContentLines(append(lines, make([]string, m.Viewport.Height())...))
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
		m.paneWidth = msg.Width
		m.Viewport.SetWidth(msg.Width)
		m.Viewport.SetHeight(max(0, msg.Height-layout.PaneChromeHeight))

		m.rebuildTitleHeader()

		if m.parsed != nil {
			m.rerender()
		}

		return nil

	case tea.BackgroundColorMsg:
		m.termBG = msg.Color

		if m.parsed != nil {
			m.rerender()
		}

		return nil
	}

	return m.Forward(msg)
}

// rerender re-renders the article at the current pane width and updates the
// viewport content. The scroll position is re-anchored to the block it was in
// rather than kept as a raw line number, so a re-render that changes block
// heights (toggling images, resizing) does not jump to unrelated content.
func (m *Model) rerender() {
	yOffset := m.Viewport.YOffset()
	oldStarts := m.blockStarts

	content, starts := m.renderArticle()
	m.blockStarts = starts

	m.setContent(content)

	maxOffset := max(0, m.ContentLines-m.Viewport.Height())
	m.Viewport.SetYOffset(min(remapYOffset(yOffset, oldStarts, starts, m.ContentLines), maxOffset))
}

// remapYOffset translates a scroll offset between two renders of the same
// blocks: the offset keeps its line position within the block it points into.
// Lines above the first block (the meta header) map unchanged, and an offset
// deeper than the block's new height clamps to the block's last line, so
// hiding a tall image leaves its label at the top of the view.
func remapYOffset(yOffset int, oldStarts, newStarts []int, newTotal int) int {
	i, found := slices.BinarySearch(oldStarts, yOffset)
	if !found {
		i--
	}

	if i < 0 || i >= len(newStarts) {
		return yOffset
	}

	// span covers the block's lines plus its trailing blank separator.
	span := newTotal + 1 - newStarts[i]
	if i+1 < len(newStarts) {
		span = newStarts[i+1] - newStarts[i]
	}

	within := yOffset - oldStarts[i]
	if within < span {
		return newStarts[i] + within
	}

	return newStarts[i] + max(0, span-2)
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

	case key.Matches(msg, m.keymap.HideImages):
		m.setShowImages(false)

	case key.Matches(msg, m.keymap.ShowImages):
		m.setShowImages(true)

	case key.Matches(msg, m.keymap.Help):
		m.showHelp = true

	case key.Matches(msg, m.keymap.OpenLink):
		return pane.OpenStoryInBrowser(m.opts.URL, m.opts.ID)

	case key.Matches(msg, m.keymap.OpenComments):
		return pane.OpenCommentsInBrowser(m.opts.ID)

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
			help.ReaderHelpScreen(m.paneWidth, m.keymap.NextStory.Enabled()),
			m.Viewport.Height(),
		)

		return header.HelpHeader("Reader Mode", m.paneWidth) + "\n" +
			content + "\n" +
			pane.FooterSeparator(m.paneWidth) + "\n" +
			help.Footer(m.paneWidth)
	}

	content := scrollbar.Attach(m.Viewport.View(), m.paneWidth, m.ContentLines, m.Viewport.Height(), m.Viewport.YOffset())

	return m.titleHeader + "\n" + content + "\n" + pane.FooterSeparator(m.paneWidth) + "\n" + m.imageIndicator()
}

// imageIndicator is the footer counterpart to the comment section's mode
// indicator, present only when the article has images for h/l to toggle.
func (m *Model) imageIndicator() string {
	if m.parsed == nil || !m.parsed.HasImages() {
		return ""
	}

	return imageStatusLine(m.showImages, m.opts.NerdFonts, m.paneWidth)
}

func imageStatusLine(show, enableNerdFonts bool, paneWidth int) string {
	icon, label := "▣", "images shown"
	if !show {
		icon, label = "▢", "images hidden"
	}

	// Nerd font glyphs render wider than one cell, so they get extra room.
	sep := " "

	if enableNerdFonts {
		icon = nerdfonts.Image
		if !show {
			icon = nerdfonts.ImageOff
		}

		sep = "  "
	}

	return xansi.Truncate("  "+icon+sep+style.Faint(label), paneWidth, "")
}

func (m *Model) rebuildTitleHeader() {
	m.titleHeader = pane.TitleHeader(m.title, m.opts.NerdFonts, layout.ReaderViewLeftMargin, m.paneWidth)
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

// setShowImages toggles image display and re-renders in place, keeping the
// scroll position. Images are always fetched, so this is instant.
func (m *Model) setShowImages(show bool) {
	if m.parsed == nil || m.showImages == show {
		return
	}

	m.showImages = show
	m.rerender()
}

// Run shows the article in a standalone reader.
func Run(parsed *article.Parsed, title string, maxWidth int, opts Options, buildHeader func(contentWidth int) string) error {
	return pane.RunStandalone(func(width, height int) pane.View {
		m := NewWithArticle(parsed, title, maxWidth, width, height, opts, buildHeader)
		m.DisableStoryNavigation()

		return m
	})
}
