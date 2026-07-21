package view

import (
	"strings"

	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/meta"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/timeago"
	"github.com/bensadeh/circumflex/view/pane"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

func (m *model) isWide() bool {
	if m.wideOverride != nil {
		return *m.wideOverride && m.width >= layout.WideViewFloor
	}

	return m.width >= max(m.config.WideViewMinWidth, layout.WideViewFloor)
}

// toggleWideLayout flips the split layout for the rest of the session,
// overriding the configured rule. Below the sanity floor there is no split
// to flip to, so the key explains itself instead of latching a no-op.
func (m *model) toggleWideLayout() tea.Cmd {
	if m.width < layout.WideViewFloor {
		return m.status.NewStatusMessageWithDuration("Terminal too narrow for the wide layout", statusMessageShort)
	}

	wide := !m.isWide()
	m.wideOverride = &wide

	m.updatePagination()
	m.resizeHelpViewport()

	// The detail view is sized to its pane, which just changed — the same
	// resize a terminal-width change delivers.
	if m.detail != nil {
		return m.detail.Update(tea.WindowSizeMsg{Width: m.detailWidth(), Height: m.height})
	}

	return nil
}

// frame is the terminal geometry for the current render; all pane sizes derive
// from it.
func (m *model) frame() layout.Frame {
	return layout.Frame{Width: m.width, Height: m.height, Wide: m.isWide()}
}

func (m *model) listWidth() int {
	return m.frame().ListWidth()
}

func (m *model) detailWidth() int {
	return m.frame().DetailWidth()
}

// detailLoading reports whether a story's comments or article are being
// fetched, as opposed to a category fetch that replaces the list itself.
func (m *model) detailLoading() bool {
	return m.fetch.detailLoading()
}

// wideDetailOpen reports whether the wide layout's detail pane is occupied:
// a story open or loading, or the help screen. The list chrome (header logo,
// page dots) dims with it, and an open story's row carries the reading
// marker that J/K move story to story.
func (m *model) wideDetailOpen() bool {
	return m.isWide() &&
		(m.detail != nil || m.screen == screenHelp || m.detailLoading())
}

func (m *model) wideView() string {
	left := paneLines(m.browsingView(), m.listWidth(), m.height)
	right := paneLines(m.detailPaneView(), m.detailWidth(), m.height)
	divider := " " + style.Faint("│") + " "

	var b strings.Builder

	for i := range m.height {
		if i > 0 {
			b.WriteByte('\n')
		}

		b.WriteString(left[i])
		b.WriteString(divider)
		b.WriteString(right[i])
	}

	return b.String()
}

func (m *model) detailPaneView() string {
	switch {
	// A link fetch keeps the article it was followed from on screen — only a
	// successful load transitions — so its feedback overlays the footer row
	// like the narrow layout's, instead of swapping in the loading pane.
	case m.fetch.linkLoading() && m.detail != nil:
		return m.overlayDetailStatus(m.detail.View(), m.detailWidth())

	// Every fetch — story, category switch, refresh, startup — spins here, so
	// loading feedback always appears in the same spot. Checked before the
	// screen so a J/K fetch spins instead of showing the outgoing story.
	case m.status.showSpinner:
		return m.loadingPane()

	case m.screen == screenHelp:
		return m.helpView()

	case m.detail != nil:
		return m.detail.View()

	default:
		return m.placeholderPane(style.Faint("Select a story"))
	}
}

// loadingPane is the detail pane while a fetch spins. A story fetch knows its
// story up front — the selection moves before the fetch starts — so the title
// heads the pane immediately, unbolded until the content arrives, over a
// dimmed placeholder for the meta block the loaded view will draw. A category
// fetch loads the list itself and has no story to name.
func (m *model) loadingPane() string {
	spinner := m.status.spinnerView()

	title := m.list.SelectedItem().Title
	if !m.detailLoading() || title == "" {
		return m.placeholderPane(spinner)
	}

	w := m.detailWidth()

	return pane.LoadingTitleHeader(title, m.config.EnableNerdFonts, layout.HeaderLeftMargin, w) +
		"\n" + m.loadingBody(w) + "\n" + m.bottomBar(w) + "\n"
}

// loadingBody fills the pane content area with the meta block placeholder
// over a centered spinner.
func (m *model) loadingBody(w int) string {
	box := m.placeholderMetaBlock(w, m.fetch.target)

	return placeholderBody(box, m.status.spinnerView(), w, m.frame().PaneContentHeight())
}

// placeholderBody stacks the meta block placeholder over content centered in
// the rest of the pane; a pane too short for both keeps just the content.
// Every line is clamped to the pane here because the error view renders its
// body outside the layouts' pane normalization, and neither the box border
// nor lipgloss wrapping keeps within panes below the wide floor.
func placeholderBody(box, content string, w, h int) string {
	contentHeight := h - lipgloss.Height(box)
	if contentHeight < 1 {
		return truncateLines(lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, content), w)
	}

	return truncateLines(box+"\n"+
		lipgloss.Place(w, contentHeight, lipgloss.Center, lipgloss.Center, content), w)
}

func truncateLines(s string, w int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = xansi.Truncate(line, w, "")
	}

	return strings.Join(lines, "\n")
}

// placeholderMetaBlock is the skeleton of the exact meta block the target
// view will draw, built through the same variant from what the list already
// knows about the selected story: the reader and the comment section lay
// theirs out at different widths, and only stories with a link get URL rows.
// target is a parameter rather than m.fetch.target because the error view
// can outlive the fetch that spawned it.
func (m *model) placeholderMetaBlock(paneWidth int, target screen) string {
	it := m.list.SelectedItem()
	d := meta.Data{
		URL:           it.URL,
		Domain:        it.Domain,
		Author:        it.Author,
		TimeAgo:       timeago.RelativeTime(it.Time),
		ID:            it.ID,
		Points:        it.Points,
		CommentsCount: it.CommentsCount,
		NerdFonts:     m.config.EnableNerdFonts,
	}

	if target == screenReader {
		skeleton := meta.ReaderMode(d).Skeleton(layout.ReaderContentWidth(paneWidth, m.config.ArticleWidth))

		return style.PrefixLines(skeleton, strings.Repeat(" ", layout.ReaderViewLeftMargin))
	}

	skeleton := meta.CommentSection(d).Skeleton(layout.CommentColumnWidth(paneWidth, m.config.CommentWidth))

	return style.PrefixLines(skeleton, strings.Repeat(" ", layout.CommentSectionLeftMargin))
}

// placeholderPane frames centered content with the same header and footer
// rules the comment section and reader draw, so opening a story fills the
// pane in place instead of popping a full frame into an empty column.
func (m *model) placeholderPane(content string) string {
	w := m.detailWidth()
	body := lipgloss.Place(w, m.frame().PaneContentHeight(), lipgloss.Center, lipgloss.Center, content)

	return "\n" + header.Underline(w) + "\n" + body + "\n" + m.bottomBar(w) + "\n"
}

// paneLines normalizes a view to exactly height rows of exactly width cells
// so panes line up when joined row by row, whatever the view emitted.
func paneLines(view string, width, height int) []string {
	lines := strings.Split(view, "\n")
	out := make([]string, height)

	for i := range out {
		var line string
		if i < len(lines) {
			line = lines[i]
		}

		switch w := xansi.StringWidth(line); {
		case w > width:
			line = xansi.Truncate(line, width, "")
		case w < width:
			line += strings.Repeat(" ", width-w)
		}

		out[i] = line
	}

	return out
}
