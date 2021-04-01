package constructor

import (
	"clx/screen"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

const (
	infoScreenText = `
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
)

func GetInfoText() string {
	longestLineLength := text.MaxLineLen(infoScreenText)
	leftOffset := screen.GetOffsetForLeftAlignedTextBlock(longestLineLength) + 5

	leftIndentation := strings.Repeat(" ", leftOffset)
	topIndentation := strings.Repeat("\n", 7)

	formattedText, _ := text.WrapWithPad(topIndentation+infoScreenText, screen.GetTerminalWidth(), leftIndentation)

	return formattedText
}
