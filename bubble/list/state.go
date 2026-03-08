package list

type ViewState int

const (
	StateStartup ViewState = iota
	StateBrowsing
	StateLoading
	StateRefreshing
	StateAddFavoritesPrompt
	StateRemoveFavoritesPrompt
	StateHelpScreen
	StateEditorOpen
)
