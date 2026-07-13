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
