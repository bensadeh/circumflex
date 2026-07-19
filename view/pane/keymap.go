package pane

import "charm.land/bubbles/v2/key"

// CommonKeyMap holds the bindings shared by both detail views; each view
// embeds it and adds its own.
type CommonKeyMap struct {
	Quit         key.Binding
	Help         key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding
	HalfPageDown key.Binding
	HalfPageUp   key.Binding
	PageDown     key.Binding
	PageUp       key.Binding
	OpenLink     key.Binding
	OpenComments key.Binding
	NextStory    key.Binding
	PrevStory    key.Binding
	ToggleWide   key.Binding

	// Search bindings apply while a search is active; n/N then take
	// precedence over each view's own jump bindings.
	Search      key.Binding
	NextMatch   key.Binding
	PrevMatch   key.Binding
	ClearSearch key.Binding
}

func DefaultCommonKeyMap() CommonKeyMap {
	return CommonKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "backspace"),
			key.WithHelp("q/esc", "back"),
		),
		Help: key.NewBinding(
			key.WithKeys("i", "?"),
			key.WithHelp("i, ?", "help"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "go to top"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "go to bottom"),
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
		OpenLink: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open story in browser"),
		),
		OpenComments: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "open comments in browser"),
		),
		NextStory: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "open next story"),
		),
		PrevStory: key.NewBinding(
			key.WithKeys("K"),
			key.WithHelp("K", "open previous story"),
		),
		ToggleWide: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "toggle wide layout"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		NextMatch: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next match"),
		),
		PrevMatch: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "prev match"),
		),
		// The same keys as Quit: views route them here first while a search
		// is active, so leaving search is one press and leaving the view a
		// second.
		ClearSearch: key.NewBinding(
			key.WithKeys("q", "esc", "backspace"),
			key.WithHelp("q/esc", "clear search"),
		),
	}
}

// DisableAppKeys removes the bindings that need the surrounding app — the
// J/K adjacent-story jumps and the z layout toggle — for standalone use
// where there is no story list and no split layout.
func (k *CommonKeyMap) DisableAppKeys() {
	k.NextStory.SetEnabled(false)
	k.PrevStory.SetEnabled(false)
	k.ToggleWide.SetEnabled(false)
}
