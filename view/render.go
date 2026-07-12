package view

import (
	"strings"

	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/view/pane"

	xansi "github.com/charmbracelet/x/ansi"

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

	if m.detail != nil {
		return m.overlayDetailStatus(m.detail.View())
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
	// Width pads but never truncates, and an error message can be wider than
	// the screen.
	lines[len(lines)-1] = xansi.Truncate(m.statusMidStyle.Width(m.width).Render(status), m.width, "")

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

// bottomBar renders the footer rule, shared with the detail views so the
// list's rule and theirs can never drift apart (memorial tint included).
func (m *model) bottomBar(width int) string {
	return pane.FooterSeparator(width)
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

	// The fixed edge slots overflow panes narrower than their sum; the center
	// can too when its width comes out non-positive and lipgloss renders the
	// message at its natural width.
	row := xansi.Truncate(left+center+right, m.listWidth(), "")

	return underline + "\n" + row
}
