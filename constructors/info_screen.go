package constructor

import (
	"clx/constants/margins"
	"clx/screen"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

const (
	infoScreenText = `
circumflex  [::d]|ˈsəːkəmflɛks|[::-]

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

func GetInfoScreen() *cview.TextView {
	helpScreen := cview.NewTextView()
	helpScreen.SetBackgroundColor(tcell.ColorDefault)
	helpScreen.SetTextColor(tcell.ColorDefault)
	helpScreen.SetTextAlign(cview.AlignLeft)
	helpScreen.SetTitleColor(tcell.ColorDefault)
	helpScreen.SetBorderColor(tcell.ColorDefault)
	helpScreen.SetTextColor(tcell.ColorDefault)
	helpScreen.SetDynamicColors(true)

	longestLineLength := 56
	lineHeight := 14

	leftOffset := screen.GetOffsetForLeftAlignedTextBlock(longestLineLength)
	topOffset := screen.GetOffsetToCenterText(lineHeight)

	helpScreen.SetPadding(topOffset-4, 0, leftOffset-margins.LeftMargin, 0)

	helpScreen.SetText(infoScreenText)

	return helpScreen
}
