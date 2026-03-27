package comments

import "charm.land/bubbles/v2/key"

// Mode represents the interaction mode for the comment view.
type Mode int

const (
	// ModeScroll is the default less-like mode where j/k scroll lines.
	ModeScroll Mode = iota
	// ModeNavigate is the comment traversal mode where j/k jump between
	// comments and h/l collapse/expand.
	ModeNavigate
)

// KeyMap defines the keybindings for the comment view.
type KeyMap struct {
	Quit       key.Binding
	ToggleMode key.Binding
	GotoTop    key.Binding
	GotoBottom key.Binding

	// Shared between modes: collapse/expand all in scroll, individual in navigate.
	NextComment key.Binding
	PrevComment key.Binding
	Collapse    key.Binding
	Expand      key.Binding
}

func defaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "back"),
		),
		ToggleMode: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle scroll/navigate mode"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "go to top"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "go to bottom"),
		),
		NextComment: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "next comment"),
		),
		PrevComment: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "prev comment"),
		),
		Collapse: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "collapse"),
		),
		Expand: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "expand"),
		),
	}
}
