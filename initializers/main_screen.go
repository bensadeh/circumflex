package builder

import (
	"clx/types"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

const (
	helpPage    = "help"
	offlinePage = "offline"
)

func NewMainView() *types.MainView {
	main := new(types.MainView)
	main.Panels = cview.NewPanels()
	main.Grid = cview.NewGrid()
	main.LeftMargin = newTextViewPrimitive("")
	main.LeftMargin.SetTextAlign(cview.AlignRight)
	main.RightMargin = newTextViewPrimitive("")
	main.Header = newTextViewPrimitive("")
	main.Footer = newTextViewPrimitive("")

	main.Grid.SetBorder(false)
	main.Grid.SetRows(2, 0, 1)
	main.Grid.SetColumns(7, 0, 3)
	main.Grid.SetBackgroundColor(tcell.ColorDefault)
	main.Grid.AddItem(main.Header, 0, 0, 1, 3, 0, 0, false)
	main.Grid.AddItem(main.Footer, 2, 0, 1, 3, 0, 0, false)
	main.Grid.AddItem(main.LeftMargin, 1, 0, 1, 1, 0, 0, false)
	main.Grid.AddItem(main.Panels, 1, 1, 1, 1, 0, 0, true)
	main.Grid.AddItem(main.RightMargin, 1, 2, 1, 1, 0, 0, false)

	main.Panels.AddPanel(helpPage, GetHelpScreen(), true, false)
	main.Panels.AddPanel(offlinePage, GetOfflineScreen(), true, false)

	return main
}

func newTextViewPrimitive(text string) *cview.TextView {
	tv := cview.NewTextView()
	tv.SetTextAlign(cview.AlignLeft)
	tv.SetText(text)
	tv.SetBorder(false)
	tv.SetBackgroundColor(tcell.ColorDefault)
	tv.SetTextColor(tcell.ColorDefault)
	tv.SetDynamicColors(true)
	return tv
}