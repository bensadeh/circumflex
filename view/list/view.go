package list

import (
	"fmt"
	"strings"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/view/list/ranking"

	"charm.land/lipgloss/v2"
)

// View renders the rank gutter and the story items — the pane's content
// between the coordinator's header and status bar.
func (m *Model) View(f Frame) string {
	content := m.contentStyle.Height(m.height).Render(m.populatedView(f))

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
		m.dimmed(f))

	return lipgloss.JoinHorizontal(lipgloss.Top, rankings, content)
}

// dimmed reports whether the pane renders in its faint "attention is
// elsewhere" treatment: while its content is being replaced (category
// switch, refresh) or while a story is open or loading next to it.
func (m *Model) dimmed(f Frame) bool {
	return m.InTransition() || f.DetailOpen
}

// storyOpen reports whether a story is open or loading in the wide layout's
// detail pane, in which case the open story's row carries the reading marker
// that J/K move story to story.
func (m *Model) storyOpen(f Frame) bool {
	return f.Wide && (f.DetailOpen || m.DetailLoading())
}

func (m *Model) PaginatorView() string {
	return m.pager.Paginator.View()
}

// DimmedPaginatorView renders every page dot faint, dropping the
// active-page marker while the list is backgrounded.
func (m *Model) DimmedPaginatorView() string {
	return m.InactiveDots(m.pager.Paginator.TotalPages)
}

// InactiveDots renders n faint page dots, used as a placeholder while the
// page count is not yet known.
func (m *Model) InactiveDots(n int) string {
	return strings.Repeat(m.styles.InactivePaginationDot.String(), n)
}

func (m *Model) populatedView(f Frame) string {
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
		m.renderItem(&b, i+start, item, f)

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
