package controller

import (
	builder "clx/initializers"
	"clx/model"
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
	MainView         *types.MainView
	SubmissionStates []*types.SubmissionState
	Category         *types.Category
}

func NewScreenController() *screenController {
	sc := new(screenController)
	sc.Application = cview.NewApplication()
	sc.SubmissionStates = []*types.SubmissionState{}
	sc.SubmissionStates = append(sc.SubmissionStates, new(types.SubmissionState))
	sc.SubmissionStates = append(sc.SubmissionStates, new(types.SubmissionState))
	sc.SubmissionStates = append(sc.SubmissionStates, new(types.SubmissionState))
	sc.SubmissionStates = append(sc.SubmissionStates, new(types.SubmissionState))
	sc.Category = new(types.Category)
	storiesToDisplay := screen.GetViewableStoriesOnSinglePage(
		screen.GetTerminalHeight(),
		maximumStoriesToDisplay)

	width := screen.GetTerminalWidth()
	height := screen.GetTerminalHeight()

	sc.SubmissionStates[types.NoCategory].MaxPages = 2
	sc.SubmissionStates[types.NoCategory].ScreenWidth = width
	sc.SubmissionStates[types.NoCategory].ScreenHeight = height
	sc.SubmissionStates[types.NoCategory].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.SubmissionStates[types.New].MaxPages = 2
	sc.SubmissionStates[types.New].ScreenWidth = width
	sc.SubmissionStates[types.New].ScreenHeight = height
	sc.SubmissionStates[types.New].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.SubmissionStates[types.Ask].MaxPages = 1
	sc.SubmissionStates[types.Ask].ScreenWidth = width
	sc.SubmissionStates[types.Ask].ScreenHeight = height
	sc.SubmissionStates[types.Ask].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.SubmissionStates[types.Show].MaxPages = 1
	sc.SubmissionStates[types.Show].ScreenWidth = width
	sc.SubmissionStates[types.Show].ScreenHeight = height
	sc.SubmissionStates[types.Show].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.MainView = builder.NewMainView()

	newsList := builder.NewList()
	sc.MainView.Panels.AddPanel(types.NewsPanel, newsList, true, false)
	sc.MainView.Panels.AddPanel(types.NewestPanel, builder.NewList(), true, false)
	sc.MainView.Panels.AddPanel(types.ShowPanel, builder.NewList(), true, false)
	sc.MainView.Panels.AddPanel(types.AskPanel, builder.NewList(), true, false)

	sc.MainView.Panels.SetCurrentPanel(types.NewsPanel)

	view.SetHackerNewsHeader(sc.MainView, width, types.NoCategory)
	view.SetLeftMarginRanks(sc.MainView, 0, storiesToDisplay)
	view.SetFooterText(sc.MainView, 0, width, 2)

	newSubs, err := model.FetchSubmissions(sc.SubmissionStates[types.NoCategory], sc.Category)

	if err != nil {
		println("Error: Could not retrieve submissions")
		os.Exit(1)
	}

	sc.SubmissionStates[types.NoCategory].Submissions = append(sc.SubmissionStates[types.NoCategory].Submissions, newSubs...)

	model.SetList(newsList, sc.SubmissionStates[types.NoCategory].Submissions, 0, storiesToDisplay, sc.Application)

	setShortcuts(sc.Application,
		sc.SubmissionStates,
		sc.MainView,
		sc.Category)

	return sc
}

func setShortcuts(app *cview.Application,
	state []*types.SubmissionState,
	main *types.MainView,
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
