package help

import (
	"clx/layout"
	"clx/nerdfonts"
	"clx/style"
	"strings"

	"charm.land/bubbles/v2/key"

	text "github.com/MichaelMure/go-term-text"
)

func MainMenuText(screenWidth int, mainMenuBindings []key.Binding) string {
	keys := new(keyList)

	for _, b := range mainMenuBindings {
		if !b.Enabled() {
			keys.addSeparator()

			continue
		}

		keys.addKeymap(b.Help().Desc, b.Help().Key)
	}

	keys.addSeparator()
	keys.addSeparator()

	return formatKeymaps(keys, screenWidth)
}

func ReaderText(screenWidth int) string {
	keys := new(keyList)

	keys.addKeymap("Down / up one line", "j, k")
	keys.addKeymap("Down / up half page", "d, u")
	keys.addSeparator()
	keys.addKeymap("Page down / up", "space/f, b")
	keys.addKeymap("Go to top / bottom", "g, G")
	keys.addKeymap("Next / prev section", "n, N")
	keys.addSeparator()
	keys.addKeymap("Help", "i, ?")
	keys.addKeymap("Back", "q, esc")

	keys.addSeparator()
	keys.addSeparator()

	return formatKeymaps(keys, screenWidth)
}

func CommentText(screenWidth int, enableNerdFonts bool) string {
	keys := new(keyList)

	keys.addHeader("Scroll Mode")
	keys.addSeparator()
	keys.addKeymap("Down / up one line", "j, k")
	keys.addKeymap("Down / up half page", "d, u")
	keys.addSeparator()
	keys.addKeymap("Page down / up", "space/f, b")
	keys.addKeymap("Go to top / bottom", "g, G")
	keys.addKeymap("Next / prev top-level comment", "n, N")
	keys.addSeparator()
	keys.addKeymap("Collapse / expand one level", "h, l")
	keys.addKeymap("Toggle collapse all", "enter")
	keys.addSeparator()
	keys.addKeymap("Switch to navigate mode", "tab")
	keys.addKeymap("Help", "i, ?")
	keys.addKeymap("Back", "q, esc")

	keys.addSeparator()

	keys.addHeader("Navigate Mode")
	keys.addSeparator()
	keys.addKeymap("Next / prev comment", "j, k")
	keys.addKeymap("Next / prev top-level comment", "n, N")
	keys.addSeparator()
	keys.addKeymap("Collapse / expand", "h, l")
	keys.addKeymap("Toggle collapse", "enter")
	keys.addSeparator()
	keys.addKeymap("Switch to scroll mode", "tab")
	keys.addKeymap("Help", "i, ?")
	keys.addKeymap("Back", "q, esc")

	keys.addSeparator()

	keys.addHeader("Legend")
	keys.addSeparator()
	keys.addKeymap("Original Poster", style.CommentOP(labelText("OP", enableNerdFonts)))
	keys.addKeymap("Grandparent Poster", style.CommentGP(labelText("GP", enableNerdFonts)))
	keys.addKeymap("Moderator", style.CommentMod(labelText("mod", enableNerdFonts)))
	keys.addSeparator()
	keys.addKeymap("New comment indicator", style.CommentNewIndicator("●"))

	keys.addSeparator()
	keys.addSeparator()

	return formatKeymaps(keys, screenWidth)
}

func formatKeymaps(keys *keyList, screenWidth int) string {
	contentWidth := min(layout.HelpScreenWidth, screenWidth-layout.HeaderLeftMargin)
	listOfKeymaps := keys.print(contentWidth)

	leftMargin := strings.Repeat(" ", layout.HeaderLeftMargin)
	output, _ := text.WrapWithPad(listOfKeymaps, screenWidth, leftMargin)

	return output
}

func labelText(fallback string, enableNerdFonts bool) string {
	if enableNerdFonts {
		return nerdfonts.Author
	}

	return fallback
}
