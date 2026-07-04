package view

import (
	"strings"

	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

const (
	// dividerWidth is the columns between the panes: a one-column rule with a
	// space of breathing room on each side.
	dividerWidth = 3

	// wideViewFloor is the narrowest terminal the split layout can render
	// sanely; below it the wide view stays off even when configured "always".
	wideViewFloor = 40
)

func (m *model) isWide() bool {
	return m.width >= max(m.config.WideViewMinWidth, wideViewFloor)
}

// listWidth is the width the list renders at: the left pane in the wide
// layout, the full screen otherwise. The divider sits in the middle of the
// screen, so both panes get an equal share.
func (m *model) listWidth() int {
	if m.isWide() {
		return (m.width - dividerWidth) / 2
	}

	return m.width
}

// detailWidth is the width the comment section and reader render at: the
// right pane in the wide layout, the full screen otherwise.
func (m *model) detailWidth() int {
	if m.isWide() {
		return m.width - m.listWidth() - dividerWidth
	}

	return m.width
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
		(m.screen == screenComments || m.screen == screenReader || m.screen == screenHelp || m.detailLoading())
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
		return m.placeholderPane(m.status.spinnerView())

	case m.screen == screenHelp:
		return m.helpView()

	case m.screen == screenReader:
		return m.readerView.View()

	case m.screen == screenComments:
		return m.commentView.View()

	default:
		return m.placeholderPane(style.Faint("Select a story"))
	}
}

// placeholderPane frames centered content with the same header and footer
// rules the comment section and reader draw, so opening a story fills the
// pane in place instead of popping a full frame into an empty column.
func (m *model) placeholderPane(content string) string {
	w := m.detailWidth()
	body := lipgloss.Place(w, m.height-headerAndFooterHeight, lipgloss.Center, lipgloss.Center, content)

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
