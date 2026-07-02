package list

import (
	"fmt"
	"strings"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/view/list/ranking"

	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	if m.state == stateHelpScreen {
		return fmt.Sprintf("%s\n%s\n%s\n%s",
			header.HelpHeader("Keyboard Shortcuts", m.width),
			m.viewport.View(),
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

	return m.listView()
}

func (m *Model) listView() string {
	var (
		sections    []string
		availHeight = m.height
	)

	v := m.titleView()
	sections = append(sections, v)
	availHeight -= lipgloss.Height(v)

	statusView := m.statusAndPaginationView()
	availHeight -= lipgloss.Height(statusView)

	content := m.contentStyle.Height(availHeight).Render(m.populatedView())
	allItems := m.VisibleItems()
	totalItems := len(allItems)

	start, end := m.pager.Paginator.GetSliceBounds(totalItems)
	readStatuses := make([]bool, end-start)

	for i, it := range allItems[start:end] {
		readStatuses[i] = m.history.Contains(it.ID)
	}

	rankings := ranking.Rankings(
		m.pager.Paginator.PerPage,
		totalItems,
		m.pager.Paginator.Page,
		m.pager.Paginator.TotalPages,
		readStatuses,
		m.dimList())

	rankingsAndContent := lipgloss.JoinHorizontal(lipgloss.Top, rankings, content)
	sections = append(sections, rankingsAndContent)
	sections = append(sections, statusView)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) titleView() string {
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
func (m *Model) bottomBar(width int) string {
	s := m.underlineStyle
	if header.MemorialActive() {
		s = s.Foreground(style.MemorialColor())
	}

	return strings.Repeat(s.Render(" "), width)
}

func (m *Model) statusAndPaginationView() string {
	var (
		centerContent string
		rightContent  string
	)

	underline := m.bottomBar(m.listWidth())

	centerContent = m.status.message

	// The page dots dim along with the list while a story is open.
	paginatorView := m.pager.Paginator.View()
	if m.wideStoryOpen() {
		paginatorView = m.dimmedPaginatorView()
	}

	switch m.state {
	case stateFetching:
		// A story fetch in the wide layout keeps the paginator so the left
		// pane doesn't change; the loading state shows in the detail pane.
		if m.isWide() && m.detailLoading() {
			rightContent = paginatorView
		} else {
			rightContent = strings.Repeat(m.styles.InactivePaginationDot.String(), m.config.PageMultiplier)
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

// dimmedPaginatorView renders every page dot faint, dropping the
// active-page marker while the list is backgrounded.
func (m *Model) dimmedPaginatorView() string {
	return strings.Repeat(m.styles.InactivePaginationDot.String(), m.pager.Paginator.TotalPages)
}

func (m *Model) populatedView() string {
	allItems := m.VisibleItems()

	var b strings.Builder

	if len(allItems) == 0 {
		if m.cat.CurrentCategory() == categories.Favorites {
			return m.favoritesEmptyMessage()
		}

		return ""
	}

	start, end := m.pager.Paginator.GetSliceBounds(len(allItems))
	itemsToShow := allItems[start:end]

	for i, item := range itemsToShow {
		m.renderItem(&b, i+start, item)

		if i != len(itemsToShow)-1 {
			fmt.Fprint(&b, strings.Repeat("\n", itemSpacing+1))
		}
	}

	// If there aren't enough items to fill up this page (always the last page)
	// then we need to add some newlines to fill up the space where items would
	// have been.
	itemsOnPage := m.pager.Paginator.ItemsOnPage(len(allItems))
	if itemsOnPage < m.pager.Paginator.PerPage {
		n := (m.pager.Paginator.PerPage - itemsOnPage) * (itemHeight + itemSpacing)

		_, _ = fmt.Fprint(&b, strings.Repeat("\n", n))
	}

	return b.String()
}

func (m *Model) favoritesEmptyMessage() string {
	dim := lipgloss.NewStyle().Faint(true)
	key := lipgloss.NewStyle().Foreground(lipgloss.Blue).Bold(true)

	// Indent to line up with where front-page item titles begin (past the rank gutter).
	margin := strings.Repeat(" ", layout.MainViewLeftMargin)

	return margin + dim.Render("No favorites yet — press ") + key.Render("f") + dim.Render(" on any story to add it")
}
