package list

import "github.com/bensadeh/circumflex/hn"

type viewState int

const (
	stateStartup viewState = iota
	stateBrowsing
	stateFetching
	stateAddFavoritesPrompt
	stateRemoveFavoritesPrompt
	stateHelpScreen
	stateReaderView
	stateCommentView
)

type transition struct {
	prevIndex int
	oldItems  []*hn.Story
	detail    bool // opening a story's comments/article rather than switching category
}
