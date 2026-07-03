package view

import (
	"fmt"
	"strings"

	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
)

func (m *model) View() string {
	if m.state == stateHelpScreen {
		return fmt.Sprintf("%s\n%s\n%s\n%s",
			header.HelpHeader("Keyboard Shortcuts", m.width),
			m.helpViewport.View(),
			m.bottomBar(m.width),
			help.Footer(m.width))
	}

	if m.isWide() {
		return m.wideView()
	}

	if m.state == stateReaderView {
		return m.readerView.View()
	}

	if m.state == stateCommentView {
		return m.commentView.View()
	}

	return m.browsingView()
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
		m.wideStoryOpen())
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

	// The page dots dim along with the list while a story is open.
	paginatorView := m.list.PaginatorView()
	if m.wideStoryOpen() {
		paginatorView = m.list.DimmedPaginatorView()
	}

	switch m.state {
	case stateFetching:
		// A story fetch in the wide layout keeps the paginator so the left
		// pane doesn't change; the loading state shows in the detail pane.
		if m.isWide() && m.detailLoading() {
			rightContent = paginatorView
		} else {
			rightContent = m.list.InactiveDots(m.config.PageMultiplier)
		}
	case stateStartup, stateBrowsing, stateAddFavoritesPrompt, stateRemoveFavoritesPrompt, stateReaderView:
		rightContent = paginatorView
	case stateCommentView:
		// Full screen, the comment view handles its own footer; in the wide
		// layout the list keeps its paginator next to it.
		if m.isWide() {
			rightContent = paginatorView
		}
	case stateHelpScreen:
		// The help screen handles its own footer.
	}

	left := m.statusLeftStyle.Render("")

	center := m.statusMidStyle.
		Width(m.listWidth() - statusBarEdgeWidth - statusBarEdgeWidth).
		Render(centerContent)

	right := m.statusEndStyle.Render(rightContent)

	return underline + "\n" + left + center + right
}
