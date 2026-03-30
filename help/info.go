package help

import (
	"clx/constants"
	"clx/keymaps"
	"clx/nerdfonts"
	"clx/style"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	text "github.com/MichaelMure/go-term-text"
)

func Text(screenWidth int, enableNerdFonts bool, mainMenuBindings []key.Binding) string {
	keys := new(keymaps.List)
	keys.Init()

	keys.AddHeader(lipgloss.NewStyle().Foreground(style.HelpMainMenuColor()).Render("Main Menu"))
	keys.AddSeparator()

	for _, b := range mainMenuBindings {
		if !b.Enabled() {
			keys.AddSeparator()

			continue
		}

		keys.AddKeymap(b.Help().Desc, b.Help().Key)
	}

	keys.AddSeparator()

	keys.AddHeader(lipgloss.NewStyle().Foreground(style.HelpCommentColor()).Render("Comment Section / Reader Mode"))
	keys.AddSeparator()
	keys.AddKeymap("Down / up one line", "j, k")
	keys.AddKeymap("Down / up one half-window", "d, u")
	keys.AddSeparator()
	keys.AddKeymap("Hide / show all replies", "h, l")
	keys.AddKeymap("Next / prev top-level comment", "n, N")
	keys.AddSeparator()
	keys.AddKeymap("Return to circumflex", "q")
	keys.AddSeparator()

	keys.AddHeader(lipgloss.NewStyle().Foreground(style.HelpLegendColor()).Render("Legend"))
	keys.AddSeparator()
	keys.AddKeymap("Original Poster", style.CommentOP(getOP(enableNerdFonts)))
	keys.AddKeymap("Grandparent Poster", style.CommentGP(getGP(enableNerdFonts)))
	keys.AddKeymap("Moderator", style.CommentMod(getMod(enableNerdFonts)))
	keys.AddSeparator()
	keys.AddKeymap("New comment indicator", style.CommentNewIndicator("●"))

	keys.AddSeparator()
	keys.AddSeparator()

	contentWidth := min(constants.HelpScreenWidth, screenWidth-constants.HeaderLeftMargin)
	listOfKeymaps := keys.Print(contentWidth)

	leftMargin := strings.Repeat(" ", constants.HeaderLeftMargin)
	output, _ := text.WrapWithPad(listOfKeymaps, screenWidth, leftMargin)

	return output
}

func getOP(enableNerdFonts bool) string {
	if enableNerdFonts {
		return nerdfonts.Author
	}

	return "OP"
}

func getGP(enableNerdFonts bool) string {
	if enableNerdFonts {
		return nerdfonts.Author
	}

	return "GP"
}

func getMod(enableNerdFonts bool) string {
	if enableNerdFonts {
		return nerdfonts.Author
	}

	return "mod"
}
