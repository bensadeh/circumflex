package view

// The display is described by orthogonal facts on the model — the screen
// that owns it, whether a fetch is in flight (fetching), and which front-page
// confirmation prompt is active — rather than one flat mode. A fetch keeps
// the screen it started on, so J/K story navigation never falls back to the
// front page while the next story loads.

type screen int

const (
	screenList screen = iota
	screenReader
	screenComments
	screenHelp
)

type prompt int

const (
	promptNone prompt = iota
	promptAddFavorite
	promptRemoveFavorite
)
