package constructor

import (
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

func GetEnvironmentScreen() *cview.TextView {
	helpScreen := cview.NewTextView()
	helpScreen.SetBackgroundColor(tcell.ColorDefault)
	helpScreen.SetTextColor(tcell.ColorDefault)
	helpScreen.SetTextAlign(cview.AlignCenter)
	helpScreen.SetTitleColor(tcell.ColorDefault)
	helpScreen.SetBorderColor(tcell.ColorDefault)
	helpScreen.SetTextColor(tcell.ColorDefault)
	helpScreen.SetDynamicColors(true)
	helpScreen.SetPadding(0, 0, -7, 0)

	helpScreen.SetText(padLines(""))

	return helpScreen
}
