package view

import (
	"github.com/bensadeh/circumflex/version"
	"github.com/bensadeh/circumflex/view/pane"
)

// windowTitle names the terminal window after what is on screen: the story
// being read, the committed query in search mode, the app itself on the front
// page. A story fetch already counts as open — the detail pane is showing its
// loading state, and a failure rolls the selection back with it.
func (m *model) windowTitle() string {
	if story := m.list.SelectedItem(); story != nil && (m.detail != nil || m.fetch.detailLoading()) {
		return pane.WindowTitle(story.Title)
	}

	if m.cat.Searching() && m.searchQuery != "" {
		return pane.WindowTitle("search: " + m.searchQuery)
	}

	return version.Name
}
