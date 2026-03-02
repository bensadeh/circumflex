package list

type ViewState int

const (
	StateStartup ViewState = iota
	StateBrowsing
	StateLoading
	StateAddFavoritesPrompt
	StateRemoveFavoritesPrompt
	StateHelpScreen
	StateEditorOpen
)
