package constructor

import (
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

const (
	offlineScreenText = `
Offline
`
)

func GetOfflineScreen() *cview.TextView {
	offlineScreen := cview.NewTextView()
	offlineScreen.SetBackgroundColor(tcell.ColorDefault)
	offlineScreen.SetTextColor(tcell.ColorDefault)
	offlineScreen.SetTextAlign(cview.AlignCenter)
	offlineScreen.SetTitle("circumflex")
	offlineScreen.SetTitleColor(tcell.ColorDefault)
	offlineScreen.SetBorderColor(tcell.ColorDefault)
	offlineScreen.SetTextColor(tcell.ColorDefault)
	offlineScreen.Box.SetBorderPadding(10, 10, 10, 10)
	offlineScreen.Box.SetBorder(true)
	offlineScreen.Box.SetBorderAttributes(tcell.AttrDim)

	offlineScreen.SetText(padLines(offlineScreenText))

	return offlineScreen
}

