package reader

import (
	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/bubbles/v2/key"
)

type keyMap struct {
	pane.CommonKeyMap

	NextHeader key.Binding
	PrevHeader key.Binding
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
	}
}
