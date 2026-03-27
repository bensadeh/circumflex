package list

import "clx/item"

type ViewState int

const (
	StateStartup ViewState = iota
	StateBrowsing
	StateFetching
	StateAddFavoritesPrompt
	StateRemoveFavoritesPrompt
	StateHelpScreen
	StateEditorOpen
	StateCommentView
)

type transition struct {
	prevIndex int
	oldItems  []*item.Story
	refresh   bool
}
