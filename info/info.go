package info

import (
	"strings"

	"clx/constants/nerdfonts"

	. "github.com/logrusorgru/aurora/v3"

	"clx/constants/margins"
	"clx/keymaps"

	text "github.com/MichaelMure/go-term-text"
)

func GetText(screenWidth int, enableNerdFonts bool) string {
	keys := new(keymaps.List)
	keys.Init()

	keys.AddHeader(Magenta(" Main Menu ").Underline().String())
	keys.AddSeparator()
	keys.AddKeymap("View comment section", "Enter")
	keys.AddKeymap("View article in Reader Mode", "Space")
	keys.AddSeparator()
	keys.AddKeymap("Refresh", "r")
	keys.AddKeymap("Change category", "Tab")
	keys.AddSeparator()
	keys.AddKeymap("Open story link in browser", "o")
	keys.AddKeymap("Open comments in browser", "c")
	keys.AddSeparator()
	keys.AddKeymap("Download article's document (only PDFs)", "d")
	keys.AddKeymap("Add to favorites", "f")
	keys.AddKeymap("Remove from favorites", "x")
	keys.AddSeparator()
	keys.AddKeymap("Bring up this screen", "i, ?")
	keys.AddKeymap("Quit to prompt", "q")
	keys.AddSeparator()

	keys.AddHeader(Yellow(" Comment Section / Reader Mode ").Underline().String())
	keys.AddSeparator()
	keys.AddKeymap("Down / up one line", "j, k")
	keys.AddKeymap("Down / up one half-window", "d, u")
	keys.AddSeparator()
	keys.AddKeymap("Hide / show all replies", "h, l")
	keys.AddKeymap("Next / prev top-level comment", "n, N")
	keys.AddSeparator()
	keys.AddKeymap("Return to circumflex", "q")
	keys.AddSeparator()

	keys.AddHeader(Blue(" Legend ").Underline().String())
	keys.AddSeparator()
	keys.AddKeymap("Original Poster", Red(getOP(enableNerdFonts)).String())
	keys.AddKeymap("Parent Poster", Magenta(getPP(enableNerdFonts)).String())
	keys.AddKeymap("Moderator", Green(getMod(enableNerdFonts)).String())
	keys.AddSeparator()
	keys.AddKeymap("New comment indicator", Cyan("●").String())

	keys.AddSeparator()
	keys.AddSeparator()

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

func getOP(enableNerdFonts bool) string {
	if enableNerdFonts {
		return nerdfonts.Author
	}

	return "OP"
}

func getPP(enableNerdFonts bool) string {
	if enableNerdFonts {
		return nerdfonts.Author
	}

	return "PP"
}

func getMod(enableNerdFonts bool) string {
	if enableNerdFonts {
		return nerdfonts.Author
	}

	return "mod"
}
