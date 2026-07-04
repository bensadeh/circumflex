// Package list is the story-list pane: a pager over fetched stories plus its
// rendering. It holds no application state — fetching, routing, the status
// bar and the split-pane layout live in the view package, which drives this
// pane through its methods and passes the per-render facts it cannot know
// itself as a Frame.
package list

import (
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/history"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/settings"

	"charm.land/bubbles/v2/paginator"
	"charm.land/lipgloss/v2"
)

// Selection is the highlight treatment of the selected row, driven by the
// coordinator's modal state (the favorites prompts).
type Selection int

const (
	SelectionNormal Selection = iota
	SelectionAddFavorite
	SelectionRemoveFavorite
)

// Frame carries the per-render facts the pane cannot know itself: how it is
// being laid out and whether a story is open or loading next to it.
type Frame struct {
	Wide          bool // rendering as the left pane of the wide layout
	DetailOpen    bool // a story's comments or article is open
	DetailLoading bool // a story's comments or article is being fetched
	Selection     Selection
}

type Model struct {
	styles     styles
	itemStyles itemStyles

	pager  pager
	width  int
	height int // rows available for list items

	config  *settings.Config
	cat     *categories.Categories
	history history.History

	// Cached style for hot-path rendering.
	contentStyle lipgloss.Style
}

func New(config *settings.Config, cat *categories.Categories, hist history.History) *Model {
	s := defaultStyles()

	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = s.ActivePaginationDot.String()
	p.InactiveDot = s.InactivePaginationDot.String()

	return &Model{
		styles:     s,
		itemStyles: newItemStyles(),
		pager: pager{
			items:     make([][]*hn.Story, categories.Count()),
			Paginator: p,
		},
		config:       config,
		cat:          cat,
		history:      hist,
		contentStyle: lipgloss.NewStyle(),
	}
}

// Resize sets the pane width and the rows available for items, and
// repaginates to fit.
func (m *Model) Resize(width, height int) {
	m.width = width
	m.height = height
	m.updatePagination()
}

func (m *Model) updatePagination() {
	// The pagination budget is one row larger than the pane: each item is
	// itemHeight+itemSpacing rows, but the last item on a page omits its
	// trailing spacing row.
	m.pager.updatePagination(m.height+1, itemHeight, itemSpacing, m.cat.CurrentCategory())
}

func (m *Model) SetItems(cat categories.Category, items []*hn.Story) {
	m.pager.items[cat] = items
}

func (m *Model) Items(cat categories.Category) []*hn.Story {
	return m.pager.items[cat]
}

func (m *Model) HasItems(cat categories.Category) bool {
	return m.pager.categoryHasStories(cat)
}

func (m *Model) CursorUp() {
	m.pager.CursorUp()
}

func (m *Model) CursorDown() {
	m.pager.CursorDown(m.cat.CurrentCategory())
}

func (m *Model) PrevPage() {
	m.pager.Paginator.PrevPage()
	m.pager.updateCursor(m.cat.CurrentCategory())
}

func (m *Model) NextPage() {
	m.pager.Paginator.NextPage()
	m.pager.updateCursor(m.cat.CurrentCategory())
}

func (m *Model) GoToTop() {
	m.pager.cursor = 0
}

func (m *Model) GoToBottom() {
	m.pager.cursor = max(0, m.pager.Paginator.ItemsOnPage(len(m.VisibleItems()))-1)
}

// ClampCursor keeps the cursor within the items on the current page, e.g.
// after the item count shrinks.
func (m *Model) ClampCursor() {
	m.pager.updateCursor(m.cat.CurrentCategory())
}

// SetIndex moves the selection to an absolute index, flipping to the page
// that contains it.
func (m *Model) SetIndex(index int) {
	m.pager.setIndex(index)
}

func (m *Model) SetPage(page int) {
	m.pager.Paginator.Page = page
}

func (m *Model) SetCursor(cursor int) {
	m.pager.cursor = cursor
}

// SetCursorClamped sets the cursor, capped to the last item on the current
// page.
func (m *Model) SetCursorClamped(cursor int) {
	itemsOnPage := m.pager.Paginator.ItemsOnPage(len(m.VisibleItems()))
	m.pager.cursor = min(cursor, itemsOnPage-1)
}

// ResetPager returns to the first page, keeping the cursor on a valid item.
func (m *Model) ResetPager() {
	currentCategory := m.cat.CurrentCategory()

	m.pager.Paginator.Page = 0
	m.pager.cursor = max(0, min(m.pager.cursor, len(m.pager.items[currentCategory])-1))
	m.updatePagination()
}

func (m *Model) Page() int {
	return m.pager.Paginator.Page
}

// Cursor is the selection's position on the current page (Index is the
// absolute position across pages).
func (m *Model) Cursor() int {
	return m.pager.cursor
}

func (m *Model) PerPage() int {
	return m.pager.Paginator.PerPage
}

// BeginTransition freezes the currently visible items so they stay on screen
// (dimmed) while replacement content is fetched. Recovering from a failed or
// cancelled fetch is the coordinator's job; it ends the transition and
// restores the category selection it captured at fetch start.
func (m *Model) BeginTransition() {
	m.pager.transition = &transition{
		oldItems: m.pager.items[m.cat.CurrentCategory()],
	}
}

func (m *Model) EndTransition() {
	m.pager.transition = nil
}

func (m *Model) InTransition() bool {
	return m.pager.transition != nil
}
