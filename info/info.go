package info

import (
	"clx/constants/messages"
	"clx/keymaps"
)

func GetStatusBarText() string {
	return messages.GetCircumflexStatusMessage()
}

func GetText(screenWidth int) string {
	return getKeymaps(screenWidth)
}

func getKeymaps(screenWidth int) string {
	keys := new(keymaps.List)
	keys.Init()

	keys.AddSeparator()
	keys.AddHeader("circumflex")
	keys.AddSeparator()
	keys.AddKeymap("Read comment section (less)", "Enter")
	keys.AddKeymap("Read article in Reader Mode (less)", "Space")
	keys.AddKeymap("Change category", "Tab")

	keys.AddKeymap("Open story link in browser", "o")
	keys.AddKeymap("Open comments in browser", "c")
	keys.AddKeymap("Refresh", "r")
	keys.AddSeparator()
	keys.AddKeymap("Add to favorites", "f")
	keys.AddKeymap("Add to favorites by ID", "F")
	keys.AddKeymap("Delete from favorites", "x")
	keys.AddSeparator()
	keys.AddKeymap("Bring up this screen", "i, ?")
	keys.AddKeymap("Quit to prompt", "q")
	keys.AddSeparator()
	keys.AddHeader("less")
	keys.AddSeparator()
	keys.AddKeymap("Down one half-window", "d")
	keys.AddKeymap("Up one half-window", "u")
	keys.AddSeparator()
	keys.AddKeymap("Jump to next top-level comment", "/ + '::'")
	keys.AddKeymap("Repeat last search", "n")
	keys.AddKeymap("Repeat last search in reverse direction", "N")
	keys.AddSeparator()
	keys.AddKeymap("Help screen", "h")
	keys.AddKeymap("Return to circumflex", "q")

	return keys.Print(screenWidth)
}
