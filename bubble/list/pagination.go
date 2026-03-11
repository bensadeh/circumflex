package list

import (
	"clx/categories"
	"clx/item"

	"charm.land/lipgloss/v2"
)

// Update pagination according to the amount of items for the current state.
func (m *Model) updatePagination() {
	index := m.Index()
	availHeight := m.height

	if m.showTitle {
		availHeight -= lipgloss.Height(m.titleView())
	}

	if m.showStatusBar {
		// We subtract one from the height because we don't want any spacing
		availHeight -= lipgloss.Height(m.statusAndPaginationView()) - 1
	}

	m.Paginator.PerPage = max(1, availHeight/(m.delegate.Height()+m.delegate.Spacing()))

	if pages := len(m.VisibleItems()); pages < 1 {
		m.Paginator.SetTotalPages(1)
	} else {
		m.Paginator.SetTotalPages(pages)
	}

	// Restore index
	m.Paginator.Page = index / m.Paginator.PerPage
	m.cursor = index % m.Paginator.PerPage

	// Make sure the page stays in bounds
	if m.Paginator.Page >= m.Paginator.TotalPages-1 {
		m.Paginator.Page = max(0, m.Paginator.TotalPages-1)
	}
}

func (m *Model) updateCursor() {
	m.cursor = min(m.cursor, m.Paginator.ItemsOnPage(len(m.VisibleItems()))-1)
}

// CursorUp moves the cursor up. This can also move the state to the previous
// page.
func (m *Model) CursorUp() {
	m.cursor--

	// If we're at the top, stop
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// CursorDown moves the cursor down. This can also advance the state to the
// next page.
func (m *Model) CursorDown() {
	itemsOnPage := m.Paginator.ItemsOnPage(len(m.VisibleItems()))

	m.cursor++

	// If we're at the end, stop
	if m.cursor < itemsOnPage {
		return
	}

	m.cursor = itemsOnPage - 1
}

// Index returns the index of the currently selected item as it appears in the
// entire slice of items.
func (m *Model) Index() int {
	return m.Paginator.Page*m.Paginator.PerPage + m.cursor
}

// Cursor returns the index of the cursor on the current page.
func (m *Model) Cursor() int {
	return m.cursor
}

// Select selects the given index of the list and goes to its respective page.
func (m *Model) Select(index int) {
	m.Paginator.Page = index / m.Paginator.PerPage
	m.cursor = index % m.Paginator.PerPage
}

// VisibleItems returns the total items available to be shown.
func (m *Model) VisibleItems() []*item.Story {
	if m.state == StateRefreshing || (m.state == StateLoading && len(m.items[categories.Buffer]) > 0) {
		return m.items[categories.Buffer]
	}

	return m.items[m.cat.GetCurrentCategory(m.favorites.HasItems())]
}

// SelectedItem returns the current selected item in the list.
func (m *Model) SelectedItem() *item.Story {
	i := m.Index()

	items := m.VisibleItems()
	if i < 0 || len(items) == 0 || len(items) <= i {
		return &item.Story{}
	}

	return items[i]
}
