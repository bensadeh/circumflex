package view

import (
	"github.com/bensadeh/circumflex/version"
	"github.com/bensadeh/circumflex/view/pane"
)

// pageTitled is the detail view that knows the title of the page it is
// showing — on a followed-link trail, not the selected story's.
type pageTitled interface {
	PageTitle() string
}

// windowTitle names the terminal window after what is on screen: the story
// being read, the committed query in search mode, the app itself on the front
// page. A story fetch already counts as open — the detail pane is showing its
// loading state, and a failure rolls the selection back with it. A detail
// view walking a trail of followed links names its current page instead.
func (m *model) windowTitle() string {
	if story := m.list.SelectedItem(); story != nil && (m.detail != nil || m.fetch.detailLoading()) {
		// The loading pane already shows the selected story's title, so it
		// wins over the outgoing view a J/K fetch leaves in the detail pane.
		if page, ok := m.detail.(pageTitled); ok && !m.fetch.detailLoading() {
			return pane.WindowTitle(page.PageTitle())
		}

		return pane.WindowTitle(story.Title)
	}

	if m.cat.Searching() && m.searchQuery != "" {
		return pane.WindowTitle("search: " + m.searchQuery)
	}

	return version.Name
}
