package controller

import (
	"clx/model"
	"clx/types"
	"clx/view"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"os"
)

const (
	helpPage    = "help"
	offlinePage = "offline"
)

func InitializeScreenController(sc *types.ScreenController) {
	storiesToDisplay := sc.SubmissionStates[types.NoCategory].ViewableStoriesOnSinglePage

	view.SetHackerNewsHeader(sc.MainView, sc.ApplicationState.ScreenWidth, types.NoCategory)
	view.SetLeftMarginRanks(sc.MainView, 0, storiesToDisplay)
	view.SetFooterText(sc.MainView, 0, sc.ApplicationState.ScreenWidth, 2)

	newSubs, err := model.FetchSubmissions(sc.SubmissionStates[types.NoCategory], sc.ApplicationState)

	if err != nil {
		println("Error: Could not retrieve submissions")
		os.Exit(1)
	}

	sc.SubmissionStates[types.NoCategory].Submissions = append(sc.SubmissionStates[types.NoCategory].Submissions, newSubs...)

	frontPanelList := model.GetListFromFrontPanel(sc.MainView.Panels)

	model.SetList(frontPanelList, sc.SubmissionStates[types.NoCategory].Submissions, 0, storiesToDisplay, sc.Application)

	setShortcuts(sc.Application,
		sc.SubmissionStates,
		sc.MainView,
		sc.ApplicationState)
}

func setShortcuts(app *cview.Application,
	submissionStates []*types.SubmissionState,
	main *types.MainView,
	appState *types.ApplicationState) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentState := submissionStates[appState.CurrentCategory]
		currentPage := currentState.CurrentPage
		viewableStories := currentState.ViewableStoriesOnSinglePage

		frontPanel, _ := main.Panels.GetFrontPanel()

		if frontPanel == offlinePage {
			if event.Rune() == 'q' {
				app.Stop()
			}
			return event
		}

		if frontPanel == helpPage {
			model.ReturnFromHelpScreen(main, appState.ScreenWidth, appState, currentPage, currentState, viewableStories)
			return event
		}

		if event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab {
			model.ChangeCategory(event, appState, submissionStates, main, app)
			return event
		} else if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			model.NextPage(app, currentState, main, appState)
		} else if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			model.PreviousPage(app, currentState, main, appState)
		} else if event.Rune() == 'q' || event.Key() == tcell.KeyEsc {
			app.Stop()
		} else if event.Rune() == 'i' || event.Rune() == '?' {
			model.ShowHelpScreen(main, appState.ScreenWidth)
		} else if event.Rune() == 'g' {
			model.SelectFirstElementInList(main)
		} else if event.Rune() == 'G' {
			model.SelectLastElementInList(currentState, main)
		}
		return event
	})
}
