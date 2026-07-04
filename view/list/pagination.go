package list

import (
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/hn"

	"charm.land/bubbles/v2/paginator"
)

// transition freezes the visible items while their replacement is fetched.
type transition struct {
	oldItems []*hn.Story
}

type pager struct {
	cursor     int
	items      [][]*hn.Story
	Paginator  paginator.Model
	transition *transition
}

func (p *pager) updatePagination(availHeight, itemHeight, itemSpacing int, currentCategory categories.Category) {
	index := p.Index()

	p.Paginator.PerPage = max(1, availHeight/(itemHeight+itemSpacing))

	p.Paginator.SetTotalPages(max(1, len(p.VisibleItems(currentCategory))))

	p.Paginator.Page = index / p.Paginator.PerPage
	p.cursor = index % p.Paginator.PerPage

	if p.Paginator.Page >= p.Paginator.TotalPages-1 {
		p.Paginator.Page = max(0, p.Paginator.TotalPages-1)
	}
}

// setIndex moves the selection to an absolute index, flipping to the page
// that contains it.
func (p *pager) setIndex(index int) {
	p.Paginator.Page = index / p.Paginator.PerPage
	p.cursor = index % p.Paginator.PerPage
}

func (p *pager) updateCursor(currentCategory categories.Category) {
	p.cursor = max(0, min(p.cursor, p.Paginator.ItemsOnPage(len(p.VisibleItems(currentCategory)))-1))
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

	p.cursor = max(0, itemsOnPage-1)
}

func (p *pager) Index() int {
	return p.Paginator.Page*p.Paginator.PerPage + p.cursor
}

func (p *pager) VisibleItems(currentCategory categories.Category) []*hn.Story {
	if p.transition != nil {
		return p.transition.oldItems
	}

	return p.items[currentCategory]
}

func (p *pager) SelectedItem(currentCategory categories.Category) *hn.Story {
	i := p.Index()

	items := p.VisibleItems(currentCategory)
	if i < 0 || len(items) == 0 || len(items) <= i {
		return &hn.Story{}
	}

	return items[i]
}

func (p *pager) categoryHasStories(cat categories.Category) bool {
	return len(p.items[cat]) != 0
}

func (m *Model) Index() int                { return m.pager.Index() }
func (m *Model) VisibleItems() []*hn.Story { return m.pager.VisibleItems(m.cat.CurrentCategory()) }
func (m *Model) SelectedItem() *hn.Story   { return m.pager.SelectedItem(m.cat.CurrentCategory()) }
