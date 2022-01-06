package constructor

import (
	"clx/constants/margins"
	"clx/constants/panels"
	"clx/core"
	"clx/favorites"
	"clx/handler"
	"clx/history"
	"clx/hn/services/hybrid"
	"clx/hn/services/mock"
	"clx/screen"
	"clx/utils/vim"

	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"
)

const (
	maximumStoriesToDisplay = 30
)

func NewScreenController(config *core.Config) *core.ScreenController {
	sc := new(core.ScreenController)
	sc.Application = cview.NewApplication()

	sc.ApplicationState = new(core.ApplicationState)
	sc.ApplicationState.StoriesToShow = screen.GetSubmissionsToShow(
		screen.GetTerminalHeight(),
		maximumStoriesToDisplay)

	sc.Articles = NewList()

	sc.MainView = NewMainView()
	sc.MainView.Panels.AddPanel(panels.StoriesPanel, sc.Articles, true, true)

	fav := favorites.Initialize()
	his := history.Initialize(config.MarkAsRead)
	sc.StoryHandler = new(handler.StoryHandler)
	sc.StoryHandler.Init(fav, his)

	sc.VimRegister = new(vim.Register)

	if config.DebugMode {
		sc.Service = new(mock.Service)
	} else {
		sc.Service = new(hybrid.Service)
		sc.Service.Init(sc.ApplicationState.StoriesToShow)
	}

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
	main.LeftMargin = newTextViewPrimitive()
	main.LeftMargin.SetTextAlign(cview.AlignRight)
	main.Header = newTextViewPrimitive()
	main.PageCounter = newTextViewPrimitive()
	main.StatusBar = newTextViewPrimitive()
	main.StatusBar.SetTextAlign(cview.AlignCenter)
	main.StatusBar.SetPadding(0, 0, -4, 0)
	main.InfoScreen = newTextViewPrimitive()

	main.Grid.SetBorder(false)
	main.Grid.SetRows(2, 0, 1)
	main.Grid.SetColumns(margins.MainViewLeftMargin, 0, margins.MainViewRightMarginPageCounter)
	main.Grid.SetBackgroundColor(tcell.ColorDefault)
	main.Grid.AddItem(main.Header, 0, 0, 1, 3, 0, 0, true)
	main.Grid.AddItem(main.LeftMargin, 1, 0, 1, 1, 0, 0, false)
	main.Grid.AddItem(main.Panels, 1, 1, 1, 2, 0, 0, false)
	main.Grid.AddItem(main.StatusBar, 2, 1, 1, 1, 0, 0, false)
	main.Grid.AddItem(main.PageCounter, 2, 2, 1, 1, 0, 0, false)

	main.Panels.AddPanel(panels.InfoPanel, main.InfoScreen, true, false)

	return main
}

func newTextViewPrimitive() *cview.TextView {
	tv := cview.NewTextView()
	tv.SetTextAlign(cview.AlignLeft)
	tv.SetBorder(false)
	tv.SetBackgroundColor(tcell.ColorDefault)
	tv.SetTextColor(tcell.ColorDefault)
	tv.SetDynamicColors(true)
	tv.SetScrollBarVisibility(cview.ScrollBarNever)

	return tv
}
