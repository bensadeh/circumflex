package pane

import (
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
)

// OpenStoryInBrowser opens the story URL, falling back to the HN item page
// for text posts. Returns nil when there is nothing to open (standalone
// reader on an arbitrary URL-less item).
func OpenStoryInBrowser(url string, id int) tea.Cmd {
	if url == "" {
		if id == 0 {
			return nil
		}

		url = hn.ItemURL(id)
	}

	return message.OpenInBrowser(url)
}

func OpenCommentsInBrowser(id int) tea.Cmd {
	if id == 0 {
		return nil
	}

	return message.OpenInBrowser(hn.ItemURL(id))
}
