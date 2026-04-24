package comments

import "charm.land/bubbles/v2/key"

type mode int

const (
	// modeRead is the default less-like mode where j/k scroll lines.
	modeRead mode = iota
	// modeNavigate is the comment traversal mode where j/k jump between
	// comments and h/l collapse/expand.
	modeNavigate
)

type keyMap struct {
	Quit         key.Binding
	ToggleMode   key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding
	NextTopLevel key.Binding
	PrevTopLevel key.Binding
	HalfPageDown key.Binding
	HalfPageUp   key.Binding
	PageDown     key.Binding
	PageUp       key.Binding

	Help key.Binding

	// Shared between modes: collapse/expand all in scroll, individual in navigate.
	NextComment    key.Binding
	PrevComment    key.Binding
	Collapse       key.Binding
	Expand         key.Binding
	ToggleCollapse key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "back"),
		),
		Help: key.NewBinding(
			key.WithKeys("i", "?"),
			key.WithHelp("i, ?", "help"),
		),
		ToggleMode: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle read/navigate mode"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "go to top"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "go to bottom"),
		),
		NextTopLevel: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next top-level comment"),
		),
		PrevTopLevel: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "prev top-level comment"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "half page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "half page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "space", "f"),
			key.WithHelp("space/f", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("b", "page up"),
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
		ToggleCollapse: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "toggle collapse"),
		),
	}
}
