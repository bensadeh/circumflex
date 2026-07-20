package comments

import (
	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/bubbles/v2/key"
)

type mode int

const (
	// modeRead is the default less-like mode where j/k scroll lines.
	modeRead mode = iota
	// modeNavigate is the comment traversal mode where j/k jump between
	// comments and h/l collapse/expand.
	modeNavigate
)

type keyMap struct {
	pane.CommonKeyMap

	NavigateMode key.Binding
	NextTopLevel key.Binding
	PrevTopLevel key.Binding

	// Shared between modes: collapse/expand all in scroll, individual in navigate.
	NextComment    key.Binding
	PrevComment    key.Binding
	Collapse       key.Binding
	Expand         key.Binding
	ToggleCollapse key.Binding

	LinkMode     key.Binding
	NextLink     key.Binding
	PrevLink     key.Binding
	OpenSelected key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		CommonKeyMap: pane.DefaultCommonKeyMap(),
		NavigateMode: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "navigate mode"),
		),
		NextTopLevel: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next top-level comment"),
		),
		PrevTopLevel: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "prev top-level comment"),
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
		LinkMode: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("⇥", "link selector"),
		),
		NextLink: key.NewBinding(
			key.WithKeys("j", "down", "n"),
			key.WithHelp("j/n", "next link"),
		),
		PrevLink: key.NewBinding(
			key.WithKeys("k", "up", "N"),
			key.WithHelp("k/N", "prev link"),
		),
		OpenSelected: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("↩", "open link"),
		),
	}
}
