package controller

import (
	"clx/model"
	"clx/types"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"unicode"
)

func SetAfterInitializationAndAfterResizeFunctions(
	app *cview.Application,
	list *cview.List,
	submissions []*types.Submissions,
	main *types.MainView,
	appState *types.ApplicationState) {
	app.SetAfterResizeFunc(func(width int, height int) {
		if appState.IsReturningFromSuspension {
			appState.IsReturningFromSuspension = false
			return
		}
		model.ResetStates(appState, submissions)
		model.InitializeHeaderAndFooterAndLeftMarginView(appState, submissions, main)
		model.FetchAndAppendSubmissionEntries(submissions[appState.CurrentCategory], appState)
		model.ShowPageAfterResize(appState, list, submissions, main, app)
		setApplicationShortcuts(app, list, submissions, main, appState)
	})
}

func setApplicationShortcuts(
	app *cview.Application,
	list *cview.List,
	submissions []*types.Submissions,
	main *types.MainView,
	appState *types.ApplicationState) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentState := submissions[appState.CurrentCategory]

		if appState.IsOnHelpScreen {
			model.ReturnFromHelpScreen(main, appState, currentState)
			return event
		}
		if event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab {
			model.ChangeCategory(event, list, appState, submissions, main, app)
			return event
		}
		if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			model.NextPage(app, list, currentState, main, appState)
			return event
		}
		if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			model.PreviousPage(app, list, currentState, main, appState)
			return event
		}
		if event.Rune() == 'q' || event.Key() == tcell.KeyEsc {
			app.Stop()
		}
		if event.Rune() == 'i' || event.Rune() == '?' {
			model.ShowHelpScreen(main, appState)
			return event
		}
		if event.Rune() == 'g' {
			model.SelectFirstElementInList(main)
			return event
		}
		if event.Rune() == 'G' {
			model.SelectLastElementInList(main, appState)
			return event
		}
		if event.Rune() == 'r' {
			afterResizeFunc := app.GetAfterResizeFunc()
			afterResizeFunc(appState.ScreenWidth, appState.ScreenHeight)
			return event
		}
		if unicode.IsDigit(event.Rune()) {
			model.SelectElementInList(main, event.Rune())
			return event
		}
		return event
	})
}
