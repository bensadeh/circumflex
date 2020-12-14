package controller

import (
	"clx/model"
	"clx/screen"
	"clx/types"
	"gitlab.com/tslocum/cview"
)

func SetAfterInitializationAndAfterResizeFunctions(app *cview.Application,
	submissionStates []*types.SubmissionState,
	main *types.MainView,
	appState *types.ApplicationState) {
	app.SetAfterResizeFunc(func(width int, height int) {
		if appState.IsReturningFromSuspension {
			appState.IsReturningFromSuspension = false
			return
		}
		resetStates(appState, submissionStates)
		model.InitializeHeaderAndFooterAndLeftMarginView(appState, submissionStates, main)
		model.FetchAndAppendSubmissions(submissionStates[appState.CurrentCategory], appState)
		model.ShowPageAfterResize(appState, submissionStates, main, app)
	})
}

func resetStates(appState *types.ApplicationState, submissionStates []*types.SubmissionState) {
	resetApplicationState(appState)
	resetSubmissionStates(submissionStates)
}

func resetApplicationState(appState *types.ApplicationState) {
	appState.CurrentPage = 0
	appState.ScreenWidth = screen.GetTerminalWidth()
	appState.ScreenHeight = screen.GetTerminalHeight()
	appState.ViewableStoriesOnSinglePage = screen.GetViewableStoriesOnSinglePage(
		appState.ScreenHeight,
		30)
}

func resetSubmissionStates(submissionStates []*types.SubmissionState) {
	numberOfCategories := 3

	for i := 0; i < numberOfCategories; i++ {
		submissionStates[i].MappedSubmissions = 0
		submissionStates[i].PageToFetchFromAPI = 0
		submissionStates[i].StoriesListed = 0
		submissionStates[i].Submissions = nil
	}
}
