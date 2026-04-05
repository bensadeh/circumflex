package list

import "github.com/bensadeh/circumflex/hn"

type ViewState int

const (
	StateStartup ViewState = iota
	StateBrowsing
	StateFetching
	StateAddFavoritesPrompt
	StateRemoveFavoritesPrompt
	StateHelpScreen
	StateReaderView
	StateCommentView
)

type transition struct {
	prevIndex int
	oldItems  []*hn.Story
	refresh   bool
}
