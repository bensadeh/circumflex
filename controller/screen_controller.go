package controller

import (
	"clx/model"
	"clx/types"
	"gitlab.com/tslocum/cview"
)

func SetResizeFunction(app *cview.Application,
	submissionStates []*types.SubmissionState,
	main *types.MainView,
	appState *types.ApplicationState) {
	app.SetAfterResizeFunc(func(width int, height int) {
		if appState.IsReturningFromSuspension {
			appState.IsReturningFromSuspension = false
			return
		}
		model.ResetStates(appState, submissionStates)
		model.InitializeHeaderAndFooterAndLeftMarginView(appState, submissionStates, main)
		model.FetchAndAppendSubmissions(submissionStates[appState.CurrentCategory], appState)
		model.ShowPageAfterResize(appState, submissionStates, main, app)
	})
}
