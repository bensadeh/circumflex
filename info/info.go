package info

import (
	"clx/constants/categories"
	"clx/constants/messages"
	"clx/keymaps"
	"clx/screen"
	"clx/settings"
	"strings"

	"github.com/gdamore/tcell/v2"

	text "github.com/MichaelMure/go-term-text"
)

const (
	numberOfCategories = 3
)

func GetStatusBarText(category int) string {
	if category == categories.Definition {
		return messages.GetCircumflexStatusMessage()
	}

	return ""
}

func GetNewCategory(event *tcell.EventKey, currentCategory int) int {
	if event.Key() == tcell.KeyBacktab {
		return getPreviousCategory(currentCategory)
	}

	return getNextCategory(currentCategory)
}

func getNextCategory(currentCategory int) int {
	isOnLastCategory := currentCategory == (numberOfCategories - 1)

	if isOnLastCategory {
		return 0
	}

	return currentCategory + 1
}

func getPreviousCategory(currentCategory int) int {
	isOnFirstCategory := currentCategory == 0

	if isOnFirstCategory {
		return numberOfCategories - 1
	}

	return currentCategory - 1
}

func GetText(category int, screenWidth int) string {
	switch category {
	case categories.Definition:
		return getDefinition()

	case categories.Keymaps:
		return getKeymaps(screenWidth)

	case categories.Settings:
		return getSettings()

	default:
		return ""
	}
}

func getDefinition() string {
	infoScreenText := `
[navy]circumflex[-::]  [::d]|ˈsəːkəmflɛks|[::-]

noun (also circumflex accent)
  a mark (^) placed over a vowel in some languages to 
  indicate contraction, length, or a particular quality.

adjective [::di]Anatomy[::-]
  bending round something else; 
  curved: [::i]circumflex coronary arteries.[::-]

[::d]ORIGIN[::-]
  late 16th century: from Latin [::bi]circumflexus[::-] 
  (from [::bi]circum[::-] ‘around, about’ + [::bi]flectere[::-] ‘to bend’), 
  translating Greek [::bi]perispōmenos[::-] ‘drawn around’.
`
	longestLineLength := text.MaxLineLen(infoScreenText)
	leftOffset := screen.GetOffsetForLeftAlignedTextBlock(longestLineLength) + 5

	leftIndentation := strings.Repeat(" ", leftOffset)
	topIndentation := strings.Repeat("\n", 7)

	formattedText, _ := text.WrapWithPad(topIndentation+infoScreenText, screen.GetTerminalWidth(), leftIndentation)

	return formattedText
}

func getKeymaps(screenWidth int) string {
	keys := new(keymaps.List)
	keys.Init()

	keys.AddSeparator()
	keys.AddHeader("Main View")
	keys.AddSeparator()
	keys.AddKeymap("Read comments", "Enter")
	keys.AddKeymap("Change category", "Tab")
	keys.AddKeymap("Open story in browser", "o")
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
	keys.AddHeader("Comment Section")
	keys.AddSeparator()
	keys.AddKeymap("Down one half-window", "d")
	keys.AddKeymap("Up one half-window", "u")
	keys.AddSeparator()
	keys.AddKeymap("Jump to next top-level comment", "/ + '::'")
	keys.AddKeymap("Repeat last search", "n")
	keys.AddKeymap("Repeat last search in reverse direction", "N")
	keys.AddSeparator()
	keys.AddKeymap("Help screen", "h")
	keys.AddKeymap("Quit to Main Screen", "q")

	return keys.Print(screenWidth)
}

func getSettings() string {
	return settings.GetSettingsText()
}
