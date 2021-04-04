package info

import (
	"clx/constants/categories"
	constructor "clx/constructors"
	"clx/screen"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

func GetText(category int) string {
	switch category {
	case categories.Definition:
		return getDefinition()

	case categories.Keymaps:
		return getKeymaps()

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

func getKeymaps() string {
	km := `     Header

Very long descriptionx
Separate item .. xyz

Add item ......... x
Delete item ...... x

     Header

Delete item ...... x
Item ......... a + b
`

	return km
}

func getSettings() string {
	return constructor.GetSettingsText()
}
