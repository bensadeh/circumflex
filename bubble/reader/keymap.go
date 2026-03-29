package reader

import "charm.land/bubbles/v2/key"

// KeyMap defines the keybindings for the reader view.
type KeyMap struct {
	Quit       key.Binding
	GotoTop    key.Binding
	GotoBottom key.Binding
	NextHeader key.Binding
	PrevHeader key.Binding
}

func defaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "back"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "go to top"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "go to bottom"),
		),
		NextHeader: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next section"),
		),
		PrevHeader: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "prev section"),
		),
	}
}
