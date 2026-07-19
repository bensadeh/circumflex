package view

import (
	"github.com/bensadeh/circumflex/help"
	"github.com/bensadeh/circumflex/view/pane"

	"charm.land/bubbles/v2/key"
)

type keyMap struct {
	Help           key.Binding
	Quit           key.Binding
	Up             key.Binding
	Down           key.Binding
	PrevPage       key.Binding
	NextPage       key.Binding
	NextCategory   key.Binding
	PrevCategory   key.Binding
	Search         key.Binding
	SearchSort     key.Binding
	SearchAge      key.Binding
	GoToTop        key.Binding
	GoToBottom     key.Binding
	OpenLink       key.Binding
	OpenComments   key.Binding
	Back           key.Binding
	Refresh        key.Binding
	AddFavorite    key.Binding
	RemoveFavorite key.Binding
	ToggleRead     key.Binding
	EnterComments  key.Binding
	ReaderMode     key.Binding
	ToggleWide     key.Binding
	Confirm        key.Binding
	Cancel         key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		EnterComments: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("↩", "View comments"),
		),
		ReaderMode: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("␣", "Reader mode"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "Refresh stories"),
		),
		NextCategory: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("⇥", "Next category"),
		),
		PrevCategory: key.NewBinding(
			key.WithKeys("shift+tab"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "Search Hacker News"),
		),
		SearchSort: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "Sort order"),
		),
		SearchAge: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "Date range"),
		),
		OpenLink: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "Open story in browser"),
		),
		OpenComments: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "Open comments in browser"),
		),
		AddFavorite: key.NewBinding(
			key.WithKeys("f", "V"),
			key.WithHelp("f", "Add favorite"),
		),
		RemoveFavorite: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "Remove favorite"),
		),
		ToggleRead: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "Toggle read"),
		),
		ToggleWide: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "Toggle wide layout"),
		),
		Help: key.NewBinding(
			key.WithKeys("i", "?"),
			key.WithHelp("i, ?", "Help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "Quit"),
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
		Back: key.NewBinding(
			key.WithKeys("backspace"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y"),
		),
		Cancel: pane.CancelKeys,
	}
}

func (km keyMap) MainMenuSections() []help.Section {
	fromBindings := func(bs ...key.Binding) []help.Item {
		items := make([]help.Item, 0, len(bs))

		for _, b := range bs {
			if item := help.FromBinding(b); item.Key != "" {
				items = append(items, item)
			}
		}

		return items
	}

	listItems := []help.Item{
		{Key: "j, k", Desc: "Down / up"},
		{Key: "h, l", Desc: "Prev / next page"},
		{Key: "g, G", Desc: "Top / bottom"},
		{Key: "⇥", Desc: "Next category"},
	}
	listItems = append(listItems, fromBindings(km.Refresh, km.AddFavorite, km.RemoveFavorite, km.ToggleRead)...)

	return []help.Section{
		{
			Title: "Open",
			Items: fromBindings(km.EnterComments, km.ReaderMode, km.OpenLink, km.OpenComments),
		},
		{
			Title: "List",
			Items: listItems,
		},
		{
			Title: "Search",
			Items: fromBindings(km.Search, km.SearchSort, km.SearchAge),
		},
		{
			Title: "App",
			Items: fromBindings(km.ToggleWide, km.Help, km.Quit),
		},
	}
}
