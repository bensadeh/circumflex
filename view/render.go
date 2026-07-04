package view

import (
	"strings"

	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
)

func (m *model) View() string {
	// In the wide layout help renders in the detail pane instead of taking
	// over the screen, so the story list stays visible next to it.
	if m.screen == screenHelp && !m.isWide() {
		return m.helpView()
	}

	if m.isWide() {
		return m.wideView()
	}

	if m.screen == screenReader {
		return m.overlayDetailStatus(m.readerView.View())
	}

	if m.screen == screenComments {
		return m.overlayDetailStatus(m.commentView.View())
	}

	return m.browsingView()
}

// overlayDetailStatus writes fetch and status feedback onto the last row of a
// full-screen detail view, which reserves that row as footer space. Narrow
// J/K story navigation stays on the open story while the next one loads, so
// its spinner and errors must surface here rather than on the front page.
func (m *model) overlayDetailStatus(view string) string {
	var status string

	switch {
	case m.fetching:
		status = m.status.spinnerView()
	case m.status.message != "":
		status = m.status.message
	default:
		return view
	}

	lines := strings.Split(view, "\n")
	lines[len(lines)-1] = m.statusMidStyle.Width(m.width).Render(status)

	return strings.Join(lines, "\n")
}

// browsingView is the front page: category header, story list, status bar.
// It fills the screen in the narrow layout and the left pane in the wide one.
func (m *model) browsingView() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.titleView(),
		m.list.View(m.listFrame()),
		m.statusAndPaginationView())
}

func (m *model) titleView() string {
	var sv string
	// In the wide layout the spinner always shows centered in the detail pane
	// instead, so loading feedback stays in one place for every kind of fetch.
	if m.status.showSpinner && !m.isWide() {
		sv = m.status.spinnerView()
	}

	return header.Header(
		m.cat.ActiveCategories(),
		m.cat.CurrentIndex(),
		m.listWidth(),
		sv,
		m.wideDetailOpen())
}

// bottomBar renders the footer rule (underlined spaces). When the HN memorial
// is active it carries the same color as the header rule (style.MemorialColor),
// so the top and bottom rules match.
func (m *model) bottomBar(width int) string {
	s := m.underlineStyle
	if header.MemorialActive() {
		s = s.Foreground(style.MemorialColor())
	}

	return strings.Repeat(s.Render(" "), width)
}

func (m *model) statusAndPaginationView() string {
	var (
		centerContent string
		rightContent  string
	)

	underline := m.bottomBar(m.listWidth())

	centerContent = m.status.message

	// The page dots dim along with the list while the detail pane is open.
	paginatorView := m.list.PaginatorView()
	if m.wideDetailOpen() {
		paginatorView = m.list.DimmedPaginatorView()
	}

	switch {
	case m.fetching:
		// A story fetch in the wide layout keeps the paginator so the left
		// pane doesn't change; the loading state shows in the detail pane.
		if m.isWide() && m.detailLoading() {
			rightContent = paginatorView
		} else {
			rightContent = m.list.InactiveDots(m.config.PageMultiplier)
		}
	case m.screen == screenComments:
		// Full screen, the comment view handles its own footer; in the wide
		// layout the list keeps its paginator next to it.
		if m.isWide() {
			rightContent = paginatorView
		}
	default:
		rightContent = paginatorView
	}

	left := m.statusLeftStyle.Render("")

	center := m.statusMidStyle.
		Width(m.listWidth() - statusBarEdgeWidth - statusBarEdgeWidth).
		Render(centerContent)

	right := m.statusEndStyle.Render(rightContent)

	return underline + "\n" + left + center + right
}
