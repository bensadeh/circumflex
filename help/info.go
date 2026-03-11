package help

import (
	"clx/constants"
	"clx/keymaps"
	"clx/nerdfonts"
	"strings"

	"charm.land/bubbles/v2/key"
	. "github.com/logrusorgru/aurora/v3"

	text "github.com/MichaelMure/go-term-text"
)

func GetText(screenWidth int, enableNerdFonts bool, mainMenuBindings []key.Binding) string {
	keys := new(keymaps.List)
	keys.Init()

	keys.AddHeader(Magenta(" Main Menu ").Underline().String())
	keys.AddSeparator()

	for _, b := range mainMenuBindings {
		if !b.Enabled() {
			keys.AddSeparator()

			continue
		}

		keys.AddKeymap(b.Help().Desc, b.Help().Key)
	}

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
	padding := screenWidth/2 - keymapsWidth/2 - constants.MainViewLeftMargin

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
