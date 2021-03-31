package constructor

import (
	"clx/constants/margins"
	"clx/constants/panels"
	"clx/core"
	"clx/screen"

	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

const (
	maximumStoriesToDisplay = 30
)

func NewScreenController() *core.ScreenController {
	sc := new(core.ScreenController)
	sc.Application = cview.NewApplication()

	sc.ApplicationState = new(core.ApplicationState)
	sc.ApplicationState.ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState.ScreenHeight = screen.GetTerminalHeight()
	sc.ApplicationState.SubmissionsToShow = screen.GetSubmissionsToShow(
		sc.ApplicationState.ScreenHeight,
		maximumStoriesToDisplay)

	sc.Articles = NewList()

	sc.MainView = NewMainView()
	sc.MainView.Panels.AddPanel(panels.SubmissionsPanel, sc.Articles, true, true)

	return sc
}

func NewList() *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.SetSelectedTextAttributes(tcell.AttrReverse)
	list.SetSelectedTextColor(tcell.ColorDefault)
	list.SetSelectedBackgroundColor(tcell.ColorDefault)
	list.SetScrollBarVisibility(cview.ScrollBarNever)

	return list
}

func NewMainView() *core.MainView {
	main := new(core.MainView)
	main.Panels = cview.NewPanels()
	main.Grid = cview.NewGrid()
	main.LeftMargin = newTextViewPrimitive("")
	main.LeftMargin.SetTextAlign(cview.AlignRight)
	main.Header = newTextViewPrimitive("")
	main.PageCounter = newTextViewPrimitive("")
	main.StatusBar = newTextViewPrimitive("")
	main.StatusBar.SetTextAlign(cview.AlignCenter)
	main.StatusBar.SetPadding(0, 0, -4, 0)
	main.Settings = newTextViewPrimitive(GetSettingsText())
	main.InfoScreen = newTextViewPrimitive(GetInfoText())

	main.Grid.SetBorder(false)
	main.Grid.SetRows(2, 0, 1)
	main.Grid.SetColumns(margins.LeftMargin, 0, margins.RightMarginPageCounter)
	main.Grid.SetBackgroundColor(tcell.ColorDefault)
	main.Grid.AddItem(main.Header, 0, 0, 1, 3, 0, 0, true)
	main.Grid.AddItem(main.LeftMargin, 1, 0, 1, 1, 0, 0, false)
	main.Grid.AddItem(main.Panels, 1, 1, 1, 2, 0, 0, false)
	main.Grid.AddItem(main.StatusBar, 2, 1, 1, 1, 0, 0, false)
	main.Grid.AddItem(main.PageCounter, 2, 2, 1, 1, 0, 0, false)

	main.Panels.AddPanel(panels.InfoPanel, main.InfoScreen, true, false)
	main.Panels.AddPanel(panels.KeymapsPanel, GetHelpScreen(), true, false)
	main.Panels.AddPanel(panels.SettingsPanel, main.Settings, true, false)

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
	tv.SetScrollBarVisibility(cview.ScrollBarNever)

	return tv
}
