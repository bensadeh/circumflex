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
		model.ShowPageAfterResize(appState, list, submissions, main)
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
			model.ChangeCategory(event, list, appState, submissions, main)
			return event
		}
		if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			model.NextPage(list, currentState, main, appState)
			return event
		}
		if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			model.PreviousPage(list, currentState, main, appState)
			return event
		}
		if event.Rune() == 'j' || event.Key() == tcell.KeyDown {
			model.SelectNextElement(list)
			return event
		}
		if event.Rune() == 'k' || event.Key() == tcell.KeyUp {
			model.SelectPreviousElement(list)
			return event
		}
		if event.Rune() == 'q' || event.Key() == tcell.KeyEsc {
			model.Quit(app)
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
			model.Refresh(app, appState)
			return event
		}
		if event.Key() == tcell.KeyEnter {
			model.ReadSubmissionComments(app, list, currentState.Entries, appState)
			return event
		}
		if event.Rune() == 'o' {
			model.OpenLinkInBrowser(list, appState, currentState.Entries)
			return event
		}
		if event.Rune() == 'c' {
			model.OpenCommentsInBrowser(list, appState, currentState.Entries)
			return event
		}
		if unicode.IsDigit(event.Rune()) {
			model.SelectElementInList(main, event.Rune())
			return event
		}
		return event
	})
}
