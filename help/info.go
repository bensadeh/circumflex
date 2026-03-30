package help

import (
	"clx/constants"
	"clx/keymaps"
	"clx/nerdfonts"
	"clx/style"
	"strings"

	"charm.land/bubbles/v2/key"

	text "github.com/MichaelMure/go-term-text"
)

func MainMenuText(screenWidth int, mainMenuBindings []key.Binding) string {
	keys := new(keymaps.List)

	for _, b := range mainMenuBindings {
		if !b.Enabled() {
			keys.AddSeparator()

			continue
		}

		keys.AddKeymap(b.Help().Desc, b.Help().Key)
	}

	keys.AddSeparator()
	keys.AddSeparator()

	return formatKeymaps(keys, screenWidth)
}

func ReaderText(screenWidth int) string {
	keys := new(keymaps.List)

	keys.AddKeymap("Down / up one line", "j, k")
	keys.AddKeymap("Down / up half page", "d, u")
	keys.AddSeparator()
	keys.AddKeymap("Page down / up", "space/f, b")
	keys.AddKeymap("Go to top / bottom", "g, G")
	keys.AddKeymap("Next / prev section", "n, N")
	keys.AddSeparator()
	keys.AddKeymap("Help", "i, ?")
	keys.AddKeymap("Back", "q, esc")

	keys.AddSeparator()
	keys.AddSeparator()

	return formatKeymaps(keys, screenWidth)
}

func CommentText(screenWidth int, enableNerdFonts bool) string {
	keys := new(keymaps.List)

	keys.AddHeader("Scroll Mode")
	keys.AddSeparator()
	keys.AddKeymap("Down / up one line", "j, k")
	keys.AddKeymap("Down / up half page", "d, u")
	keys.AddSeparator()
	keys.AddKeymap("Page down / up", "space/f, b")
	keys.AddKeymap("Go to top / bottom", "g, G")
	keys.AddKeymap("Next / prev top-level comment", "n, N")
	keys.AddSeparator()
	keys.AddKeymap("Collapse / expand one level", "h, l")
	keys.AddKeymap("Toggle collapse all", "enter")
	keys.AddSeparator()
	keys.AddKeymap("Switch to navigate mode", "tab")
	keys.AddKeymap("Help", "i, ?")
	keys.AddKeymap("Back", "q, esc")

	keys.AddSeparator()

	keys.AddHeader("Navigate Mode")
	keys.AddSeparator()
	keys.AddKeymap("Next / prev comment", "j, k")
	keys.AddKeymap("Next / prev top-level comment", "n, N")
	keys.AddSeparator()
	keys.AddKeymap("Collapse / expand", "h, l")
	keys.AddKeymap("Toggle collapse", "enter")
	keys.AddSeparator()
	keys.AddKeymap("Switch to scroll mode", "tab")
	keys.AddKeymap("Help", "i, ?")
	keys.AddKeymap("Back", "q, esc")

	keys.AddSeparator()

	keys.AddHeader("Legend")
	keys.AddSeparator()
	keys.AddKeymap("Original Poster", style.CommentOP(labelText("OP", enableNerdFonts)))
	keys.AddKeymap("Grandparent Poster", style.CommentGP(labelText("GP", enableNerdFonts)))
	keys.AddKeymap("Moderator", style.CommentMod(labelText("mod", enableNerdFonts)))
	keys.AddSeparator()
	keys.AddKeymap("New comment indicator", style.CommentNewIndicator("●"))

	keys.AddSeparator()
	keys.AddSeparator()

	return formatKeymaps(keys, screenWidth)
}

func formatKeymaps(keys *keymaps.List, screenWidth int) string {
	contentWidth := min(constants.HelpScreenWidth, screenWidth-constants.HeaderLeftMargin)
	listOfKeymaps := keys.Print(contentWidth)

	leftMargin := strings.Repeat(" ", constants.HeaderLeftMargin)
	output, _ := text.WrapWithPad(listOfKeymaps, screenWidth, leftMargin)

	return output
}

func labelText(fallback string, enableNerdFonts bool) string {
	if enableNerdFonts {
		return nerdfonts.Author
	}

	return fallback
}
