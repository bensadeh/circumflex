package constructor

import (
	"clx/screen"
	"clx/types"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

const (
	maximumStoriesToDisplay = 30
	helpPage                = "help"
	offlinePage             = "offline"
)

func NewScreenController() *types.ScreenController {
	sc := new(types.ScreenController)

	sc.Application = cview.NewApplication()

	sc.SubmissionStates = []*types.SubmissionState{}
	sc.SubmissionStates = append(sc.SubmissionStates, new(types.SubmissionState))
	sc.SubmissionStates = append(sc.SubmissionStates, new(types.SubmissionState))
	sc.SubmissionStates = append(sc.SubmissionStates, new(types.SubmissionState))
	sc.SubmissionStates = append(sc.SubmissionStates, new(types.SubmissionState))

	sc.ApplicationState = new(types.ApplicationState)
	sc.ApplicationState.ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState.ScreenHeight = screen.GetTerminalHeight()

	storiesToDisplay := screen.GetViewableStoriesOnSinglePage(sc.ApplicationState.ScreenHeight, maximumStoriesToDisplay)

	sc.SubmissionStates[types.NoCategory].MaxPages = 2
	sc.SubmissionStates[types.NoCategory].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.SubmissionStates[types.New].MaxPages = 2
	sc.SubmissionStates[types.New].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.SubmissionStates[types.Ask].MaxPages = 1
	sc.SubmissionStates[types.Ask].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.SubmissionStates[types.Show].MaxPages = 1
	sc.SubmissionStates[types.Show].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.MainView = NewMainView()

	newsList := NewList()
	sc.MainView.Panels.AddPanel(types.NewsPanel, newsList, true, false)
	sc.MainView.Panels.AddPanel(types.NewestPanel, NewList(), true, false)
	sc.MainView.Panels.AddPanel(types.ShowPanel, NewList(), true, false)
	sc.MainView.Panels.AddPanel(types.AskPanel, NewList(), true, false)

	sc.MainView.Panels.SetCurrentPanel(types.NewsPanel)

	return sc
}

func NewList() *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)

	return list
}

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
