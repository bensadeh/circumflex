package info

import (
	"clx/constants/margins"
	"clx/constants/messages"
	"clx/keymaps"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

func GetStatusBarText() string {
	return messages.GetCircumflexStatusMessage()
}

func GetText(screenWidth int) string {
	keys := new(keymaps.List)
	keys.Init()

	keys.AddHeader("circumflex")
	keys.AddSeparator()
	keys.AddKeymap("Read comment section (less)", "Enter")
	keys.AddKeymap("Read article in Reader Mode (less)", "Space")
	keys.AddKeymap("Force Read article in Reader Mode (less)", "t")
	keys.AddSeparator()
	keys.AddKeymap("Refresh", "r")
	keys.AddKeymap("Change category", "Tab")
	keys.AddSeparator()
	keys.AddKeymap("Open story link in browser", "o")
	keys.AddKeymap("Open comments in browser", "c")
	keys.AddSeparator()
	keys.AddKeymap("Add to favorites", "f")
	keys.AddKeymap("Remove from favorites", "x")
	keys.AddSeparator()
	keys.AddKeymap("Bring up this screen", "i, ?")
	keys.AddKeymap("Quit to prompt", "q")
	keys.AddSeparator()
	keys.AddHeader("less")
	keys.AddSeparator()
	keys.AddKeymap("Down one half-window", "d")
	keys.AddKeymap("Up one half-window", "u")
	keys.AddSeparator()
	keys.AddKeymap("Jump to next top-level comment", "n")
	keys.AddKeymap("Jump to previous top-level comment", "N")
	keys.AddSeparator()
	keys.AddKeymap("Help screen", "h")
	keys.AddKeymap("Return to circumflex", "q")

	keymapsWidth := 80
	listOfKeymaps := keys.Print(keymapsWidth)
	listOfKeymapsCentered := alignCenter(listOfKeymaps, screenWidth, keymapsWidth)

	return listOfKeymapsCentered
}

func alignCenter(input string, screenWidth int, keymapsWidth int) string {
	padding := screenWidth/2 - keymapsWidth/2 - margins.MainViewLeftMargin

	if padding < 0 {
		return input
	}

	padToCenterAlign := strings.Repeat(" ", padding)
	output, _ := text.WrapWithPad(input, screenWidth, padToCenterAlign)

	return output
}
