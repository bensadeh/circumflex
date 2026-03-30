package list

import (
	"clx/categories"
	"clx/item"

	"charm.land/bubbles/v2/paginator"
	"charm.land/lipgloss/v2"
)

type pager struct {
	cursor     int
	items      [][]*item.Story
	Paginator  paginator.Model
	transition *transition
}

func (p *pager) updatePagination(availHeight, itemHeight, itemSpacing int, currentCategory categories.Category) {
	index := p.Index()

	p.Paginator.PerPage = max(1, availHeight/(itemHeight+itemSpacing))

	if pages := len(p.VisibleItems(currentCategory)); pages < 1 {
		p.Paginator.SetTotalPages(1)
	} else {
		p.Paginator.SetTotalPages(pages)
	}

	// Restore index
	p.Paginator.Page = index / p.Paginator.PerPage
	p.cursor = index % p.Paginator.PerPage

	// Make sure the page stays in bounds
	if p.Paginator.Page >= p.Paginator.TotalPages-1 {
		p.Paginator.Page = max(0, p.Paginator.TotalPages-1)
	}
}

func (m *Model) updatePagination() {
	availHeight := m.height
	availHeight -= lipgloss.Height(m.titleView())

	// We subtract one from the height because we don't want any spacing
	availHeight -= lipgloss.Height(m.statusAndPaginationView()) - 1

	m.pager.updatePagination(availHeight, m.delegate.Height(), m.delegate.Spacing(), m.cat.CurrentCategory())
}

func (p *pager) updateCursor(currentCategory categories.Category) {
	p.cursor = min(p.cursor, p.Paginator.ItemsOnPage(len(p.VisibleItems(currentCategory)))-1)
}

func (p *pager) CursorUp() {
	p.cursor--

	if p.cursor < 0 {
		p.cursor = 0
	}
}

func (p *pager) CursorDown(currentCategory categories.Category) {
	itemsOnPage := p.Paginator.ItemsOnPage(len(p.VisibleItems(currentCategory)))

	p.cursor++

	if p.cursor < itemsOnPage {
		return
	}

	p.cursor = itemsOnPage - 1
}

func (p *pager) Index() int {
	return p.Paginator.Page*p.Paginator.PerPage + p.cursor
}

func (p *pager) Cursor() int {
	return p.cursor
}

func (p *pager) Select(index int) {
	p.Paginator.Page = index / p.Paginator.PerPage
	p.cursor = index % p.Paginator.PerPage
}

func (p *pager) VisibleItems(currentCategory categories.Category) []*item.Story {
	if p.transition != nil {
		return p.transition.oldItems
	}

	return p.items[currentCategory]
}

func (p *pager) SelectedItem(currentCategory categories.Category) *item.Story {
	i := p.Index()

	items := p.VisibleItems(currentCategory)
	if i < 0 || len(items) == 0 || len(items) <= i {
		return &item.Story{}
	}

	return items[i]
}

func (p *pager) categoryHasStories(cat categories.Category) bool {
	return len(p.items[cat]) != 0
}

func (m *Model) Index() int                  { return m.pager.Index() }
func (m *Model) VisibleItems() []*item.Story { return m.pager.VisibleItems(m.cat.CurrentCategory()) }
func (m *Model) SelectedItem() *item.Story   { return m.pager.SelectedItem(m.cat.CurrentCategory()) }
