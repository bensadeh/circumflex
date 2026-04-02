package list

import (
	"testing"

	"github.com/bensadeh/circumflex/item"

	"charm.land/bubbles/v2/paginator"
	"github.com/stretchr/testify/assert"
)

func newTestPager(itemCount, perPage int) pager {
	p := paginator.New()
	p.PerPage = perPage

	stories := make([]*item.Story, itemCount)
	for i := range stories {
		stories[i] = &item.Story{ID: i + 1}
	}

	items := make([][]*item.Story, numberOfCategories)
	items[0] = stories

	pg := pager{
		items:     items,
		Paginator: p,
	}

	pg.Paginator.SetTotalPages(len(stories))

	return pg
}

func TestPager_Index(t *testing.T) {
	p := newTestPager(30, 10)

	assert.Equal(t, 0, p.Index())

	p.cursor = 5
	assert.Equal(t, 5, p.Index())

	p.Paginator.Page = 1
	p.cursor = 3
	assert.Equal(t, 13, p.Index())
}

func TestPager_CursorUp(t *testing.T) {
	p := newTestPager(10, 5)
	p.cursor = 3

	p.CursorUp()
	assert.Equal(t, 2, p.cursor)

	p.CursorUp()
	assert.Equal(t, 1, p.cursor)

	// Clamps at zero
	p.cursor = 0
	p.CursorUp()
	assert.Equal(t, 0, p.cursor)
}

func TestPager_CursorDown(t *testing.T) {
	p := newTestPager(10, 5)

	p.CursorDown(0)
	assert.Equal(t, 1, p.cursor)

	p.CursorDown(0)
	assert.Equal(t, 2, p.cursor)

	// Clamps at last item on page
	p.cursor = 4
	p.CursorDown(0)
	assert.Equal(t, 4, p.cursor)
}

func TestPager_Select(t *testing.T) {
	p := newTestPager(30, 10)

	p.Select(15)
	assert.Equal(t, 1, p.Paginator.Page)
	assert.Equal(t, 5, p.cursor)

	p.Select(0)
	assert.Equal(t, 0, p.Paginator.Page)
	assert.Equal(t, 0, p.cursor)
}

func TestPager_VisibleItems(t *testing.T) {
	p := newTestPager(5, 10)

	items := p.VisibleItems(0)
	assert.Len(t, items, 5)
}

func TestPager_VisibleItems_WithTransition(t *testing.T) {
	p := newTestPager(5, 10)
	oldItems := []*item.Story{{ID: 99}}
	p.transition = &transition{oldItems: oldItems}

	items := p.VisibleItems(0)
	assert.Len(t, items, 1)
	assert.Equal(t, 99, items[0].ID)
}

func TestPager_SelectedItem(t *testing.T) {
	p := newTestPager(5, 10)
	p.cursor = 2

	selected := p.SelectedItem(0)
	assert.Equal(t, 3, selected.ID)
}

func TestPager_SelectedItem_OutOfBounds(t *testing.T) {
	p := newTestPager(0, 10)

	selected := p.SelectedItem(0)
	assert.Equal(t, 0, selected.ID)
}

func TestPager_CategoryHasStories(t *testing.T) {
	p := newTestPager(5, 10)

	assert.True(t, p.categoryHasStories(0))
	assert.False(t, p.categoryHasStories(1))
}

func TestPager_UpdatePagination(t *testing.T) {
	p := newTestPager(25, 1)

	p.updatePagination(30, 2, 1, 0)
	assert.Equal(t, 10, p.Paginator.PerPage)
	assert.Equal(t, 3, p.Paginator.TotalPages)
}

func TestPager_UpdatePagination_RestoresIndex(t *testing.T) {
	p := newTestPager(20, 5)
	p.Paginator.SetTotalPages(20)
	p.Select(12)

	p.updatePagination(30, 2, 1, 0)

	assert.Equal(t, 12, p.Index())
}

func TestPager_UpdateCursor(t *testing.T) {
	p := newTestPager(3, 10)
	p.cursor = 5

	p.updateCursor(0)
	assert.Equal(t, 2, p.cursor)
}

func TestPager_Cursor(t *testing.T) {
	p := newTestPager(10, 5)

	assert.Equal(t, 0, p.Cursor())

	p.cursor = 3
	assert.Equal(t, 3, p.Cursor())
}

func TestPager_CategoryHasStories_MultipleCategories(t *testing.T) {
	p := newTestPager(5, 10)
	p.items[2] = []*item.Story{{ID: 1}}
	p.items[3] = []*item.Story{}

	assert.True(t, p.categoryHasStories(0))
	assert.False(t, p.categoryHasStories(1))
	assert.True(t, p.categoryHasStories(2))
	assert.False(t, p.categoryHasStories(3))
	assert.False(t, p.categoryHasStories(4))
}
