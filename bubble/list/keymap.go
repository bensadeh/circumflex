package list

import "charm.land/bubbles/v2/key"

// KeyMap defines all keybindings for the list view.
type KeyMap struct {
	Help           key.Binding
	Quit           key.Binding
	Up             key.Binding
	Down           key.Binding
	PrevPage       key.Binding
	NextPage       key.Binding
	NextCategory   key.Binding
	PrevCategory   key.Binding
	GoToTop        key.Binding
	GoToBottom     key.Binding
	OpenLink       key.Binding
	OpenComments   key.Binding
	Refresh        key.Binding
	AddFavorite    key.Binding
	RemoveFavorite key.Binding
	EnterComments  key.Binding
	ReaderMode     key.Binding
	Confirm        key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		EnterComments: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "View comment section"),
		),
		ReaderMode: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("Space", "View article in Reader Mode"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "Refresh"),
		),
		NextCategory: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("Tab", "Change category"),
		),
		PrevCategory: key.NewBinding(
			key.WithKeys("shift+tab"),
		),
		OpenLink: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "Open story link in browser"),
		),
		OpenComments: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "Open comments in browser"),
		),
		AddFavorite: key.NewBinding(
			key.WithKeys("f", "V"),
			key.WithHelp("f", "Add to favorites"),
		),
		RemoveFavorite: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "Remove from favorites"),
		),
		Help: key.NewBinding(
			key.WithKeys("i", "?"),
			key.WithHelp("i, ?", "Bring up this screen"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "Quit to prompt"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("left", "h"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right", "l"),
		),
		GoToTop: key.NewBinding(
			key.WithKeys("g"),
		),
		GoToBottom: key.NewBinding(
			key.WithKeys("G"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y"),
		),
	}
}

// MainMenuBindings returns the bindings shown in the help screen's "Main Menu"
// section. Zero-value bindings act as group separators.
func (km KeyMap) MainMenuBindings() []key.Binding {
	sep := key.Binding{}
	return []key.Binding{
		km.EnterComments,
		km.ReaderMode,
		sep,
		km.Refresh,
		km.NextCategory,
		sep,
		km.OpenLink,
		km.OpenComments,
		sep,
		km.AddFavorite,
		km.RemoveFavorite,
		sep,
		km.Help,
		km.Quit,
	}
}
