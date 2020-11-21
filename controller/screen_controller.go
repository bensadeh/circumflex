package controller

import (
	"clx/model"
	"clx/primitives"
	"clx/screen"
	"clx/types"
	"clx/view"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"os"
)

const (
	maximumStoriesToDisplay = 30
	helpPage                = "help"
	offlinePage             = "offline"
)

type screenController struct {
	Application      *cview.Application
	MainView         *primitives.MainView
	ApplicationState []*types.ApplicationState
	Category         *types.Category
}

func NewScreenController() *screenController {
	sc := new(screenController)
	sc.Application = cview.NewApplication()
	sc.ApplicationState = []*types.ApplicationState{}
	sc.ApplicationState = append(sc.ApplicationState, new(types.ApplicationState))
	sc.ApplicationState = append(sc.ApplicationState, new(types.ApplicationState))
	sc.ApplicationState = append(sc.ApplicationState, new(types.ApplicationState))
	sc.ApplicationState = append(sc.ApplicationState, new(types.ApplicationState))
	sc.Category = new(types.Category)
	storiesToDisplay := screen.GetViewableStoriesOnSinglePage(
		screen.GetTerminalHeight(),
		maximumStoriesToDisplay)

	width := screen.GetTerminalWidth()
	height := screen.GetTerminalHeight()

	sc.ApplicationState[types.NoCategory].MaxPages = 2
	sc.ApplicationState[types.NoCategory].ScreenWidth = width
	sc.ApplicationState[types.NoCategory].ScreenHeight = height
	sc.ApplicationState[types.NoCategory].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.ApplicationState[types.New].MaxPages = 2
	sc.ApplicationState[types.New].ScreenWidth = width
	sc.ApplicationState[types.New].ScreenHeight = height
	sc.ApplicationState[types.New].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.ApplicationState[types.Ask].MaxPages = 1
	sc.ApplicationState[types.Ask].ScreenWidth = width
	sc.ApplicationState[types.Ask].ScreenHeight = height
	sc.ApplicationState[types.Ask].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.ApplicationState[types.Show].MaxPages = 1
	sc.ApplicationState[types.Show].ScreenWidth = width
	sc.ApplicationState[types.Show].ScreenHeight = height
	sc.ApplicationState[types.Show].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.MainView = primitives.NewMainView(width, storiesToDisplay)

	newsList := createNewList()
	sc.MainView.Panels.AddPanel(types.NewsPanel, newsList, true, false)
	sc.MainView.Panels.AddPanel(types.NewestPanel, createNewList(), true, false)
	sc.MainView.Panels.AddPanel(types.ShowPanel, createNewList(), true, false)
	sc.MainView.Panels.AddPanel(types.AskPanel, createNewList(), true, false)

	sc.MainView.Panels.SetCurrentPanel(types.NewsPanel)

	view.SetHackerNewsHeader(sc.MainView, width, types.NoCategory)
	view.SetLeftMarginRanks(sc.MainView, 0, storiesToDisplay)
	view.SetFooterText(sc.MainView, 0, width, 2)

	newSubs, err := model.FetchSubmissions(sc.ApplicationState[types.NoCategory], sc.Category)

	if err != nil {
		println("Error: Could not retrieve submissions")
		os.Exit(1)
	}

	sc.ApplicationState[types.NoCategory].Submissions = append(sc.ApplicationState[types.NoCategory].Submissions, newSubs...)

	model.SetList(newsList, sc.ApplicationState[types.NoCategory].Submissions, 0, storiesToDisplay, sc.Application)

	setShortcuts(sc.Application,
		sc.ApplicationState,
		sc.MainView,
		sc.Category)

	return sc
}

func setShortcuts(app *cview.Application,
	state []*types.ApplicationState,
	main *primitives.MainView,
	cat *types.Category) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentState := state[cat.CurrentCategory]
		currentPage := currentState.CurrentPage
		screenWidth := currentState.ScreenWidth
		viewableStories := currentState.ViewableStoriesOnSinglePage

		frontPanel, _ := main.Panels.GetFrontPanel()

		if frontPanel == offlinePage {
			if event.Rune() == 'q' {
				app.Stop()
			}
			return event
		}

		if frontPanel == helpPage {
			model.ReturnFromHelpScreen(main, screenWidth, cat, currentPage, currentState, viewableStories)
			return event
		}

		if event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab {
			model.ChangeCategory(event, cat, state, main, app)
			return event
		} else if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			model.NextPage(app, currentState, main, cat)
		} else if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			model.PreviousPage(app, currentState, main, main.Panels)
		} else if event.Rune() == 'q' || event.Key() == tcell.KeyEsc {
			app.Stop()
		} else if event.Rune() == 'i' || event.Rune() == '?' {
			model.ShowHelpScreen(main, screenWidth)
		} else if event.Rune() == 'g' {
			model.SelectFirstElementInList(main)
		} else if event.Rune() == 'G' {
			model.SelectLastElementInList(currentState, main)
		}
		return event
	})
}

func createNewList() *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)

	return list
}
