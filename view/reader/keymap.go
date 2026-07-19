package reader

import (
	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/bubbles/v2/key"
)

type keyMap struct {
	pane.CommonKeyMap

	NextHeader key.Binding
	PrevHeader key.Binding
	HideImages key.Binding
	ShowImages key.Binding

	LinkMode     key.Binding
	NextLink     key.Binding
	PrevLink     key.Binding
	JumpNextLink key.Binding
	JumpPrevLink key.Binding
	OpenSelected key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		CommonKeyMap: pane.DefaultCommonKeyMap(),
		NextHeader: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next section"),
		),
		PrevHeader: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "prev section"),
		),
		HideImages: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "hide images"),
		),
		ShowImages: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "show images"),
		),
		LinkMode: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("⇥", "link selector"),
		),
		NextLink: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j", "next link on screen"),
		),
		PrevLink: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k", "prev link on screen"),
		),
		JumpNextLink: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "jump to next link"),
		),
		JumpPrevLink: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "jump to prev link"),
		),
		OpenSelected: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("↩", "open link"),
		),
	}
}
