package controller

import (
	"clx/model"
	"clx/screen"
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

func setShortcuts(app *cview.Application,
	submissionStates []*types.SubmissionState,
	main *types.MainView,
	appState *types.ApplicationState) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentState := submissionStates[appState.CurrentCategory]

		frontPanel, _ := main.Panels.GetFrontPanel()

		if frontPanel == offlinePage {
			if event.Rune() == 'q' {
				app.Stop()
			}
			return event
		}

		if frontPanel == helpPage {
			model.ReturnFromHelpScreen(main, appState, currentState)
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
			model.ShowHelpScreen(main, appState)
		} else if event.Rune() == 'g' {
			model.SelectFirstElementInList(main)
		} else if event.Rune() == 'G' {
			model.SelectLastElementInList(main, appState)
		}
		return event
	})
}

func SetResizeFunction(app *cview.Application,
	submissionStates []*types.SubmissionState,
	main *types.MainView,
	appState *types.ApplicationState) {
	app.SetAfterResizeFunc(func(width int, height int) {
		if appState.IsReturningFromSuspension {
			appState.IsReturningFromSuspension = false
			return
		}

		appState.CurrentPage = 0
		appState.ScreenWidth = screen.GetTerminalWidth()
		appState.ScreenHeight = screen.GetTerminalHeight()
		appState.ViewableStoriesOnSinglePage = screen.GetViewableStoriesOnSinglePage(
			appState.ScreenHeight,
			30)

		ClearSubmissionStates(submissionStates)

		view.SetPanelCategory(main, appState.CurrentCategory)
		view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.CurrentCategory)
		view.SetLeftMarginRanks(main, 0, appState.ViewableStoriesOnSinglePage)
		view.SetFooterText(main,
			0,
			appState.ScreenWidth,
			submissionStates[appState.CurrentCategory].MaxPages)

		newSubs, err := model.FetchSubmissions(submissionStates[appState.CurrentCategory], appState)

		if err != nil {
			println("Error: Could not retrieve submissions")
			os.Exit(1)
		}

		submissionStates[appState.CurrentCategory].Submissions = append(submissionStates[appState.CurrentCategory].Submissions, newSubs...)

		frontPanelList := model.GetListFromFrontPanel(main.Panels)

		model.SetList(frontPanelList,
			submissionStates[appState.CurrentCategory].Submissions,
			appState,
			app)

		if appState.IsOnHelpScreen {
			model.ShowHelpScreen(main, appState)
		}

		setShortcuts(app, submissionStates, main, appState)
	})
}

func ClearSubmissionStates(submissionStates []*types.SubmissionState) {
	numberOfCategories := 3

	for i := 0; i < numberOfCategories; i++ {
		submissionStates[i].MappedSubmissions = 0
		submissionStates[i].PageToFetchFromAPI = 0
		submissionStates[i].StoriesListed = 0
		submissionStates[i].Submissions = nil
	}
}
