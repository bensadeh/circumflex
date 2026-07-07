package view

import (
	"strings"

	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

func (m *model) isWide() bool {
	return m.width >= max(m.config.WideViewMinWidth, layout.WideViewFloor)
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
	return m.fetching && m.detailFetch
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
// heads the pane immediately, unbolded until the content arrives. A category
// fetch loads the list itself and has no story to name.
func (m *model) loadingPane() string {
	spinner := m.status.spinnerView()

	title := m.list.SelectedItem().Title
	if !m.detailLoading() || title == "" {
		return m.placeholderPane(spinner)
	}

	w := m.detailWidth()
	body := lipgloss.Place(w, m.frame().PaneContentHeight(), lipgloss.Center, lipgloss.Center, spinner)

	return pane.LoadingTitleHeader(title, m.config.EnableNerdFonts, layout.HeaderLeftMargin, w) +
		"\n" + body + "\n" + m.bottomBar(w) + "\n"
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
