package view

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
