package controller

import (
	"clx/model"
	"clx/structs"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"unicode"
)

func SetAfterInitializationAndAfterResizeFunctions(
	app *cview.Application,
	list *cview.List,
	submissions []*structs.Submissions,
	main *structs.MainView,
	appState *structs.ApplicationState) {
	model.SetAfterInitializationAndAfterResizeFunctions(app, list, submissions, main, appState)
}

func SetApplicationShortcuts(
	app *cview.Application,
	list *cview.List,
	submissions []*structs.Submissions,
	main *structs.MainView,
	appState *structs.ApplicationState) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentState := submissions[appState.SubmissionsCategory]

		//Offline events
		if appState.IsOffline && event.Rune() == 'r' {
			model.Refresh(app, list, main, submissions, appState)
			return event
		}
		if appState.IsOffline && event.Rune() == 'q' {
			model.Quit(app)
			return event
		}
		if appState.IsOffline {
			return event
		}

		//Help screen events
		if appState.IsOnHelpScreen && (event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab) {
			model.ChangeHelpScreenCategory(event, appState, main)
			return event
		}
		if appState.IsOnHelpScreen && (event.Rune() == 'i' || event.Rune() == 'q'){
			model.ExitHelpScreen(main, appState, currentState)
			return event
		}
		if appState.IsOnHelpScreen {
			return event
		}

		//Submission screen events
		if event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab {
			model.ChangeCategory(app, event, list, appState, submissions, main)
			return event
		}
		if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			model.NextPage(app, list, currentState, main, appState)
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
			model.EnterInfoScreen(main, appState)
			return event
		}
		if event.Rune() == 'g' {
			model.SelectFirstElementInList(list)
			return event
		}
		if event.Rune() == 'G' {
			model.SelectLastElementInList(list)
			return event
		}
		if event.Rune() == 'r' {
			model.Refresh(app, list, main, submissions, appState)
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
			model.SelectElementInList(list, event.Rune())
			return event
		}
		return event
	})
}
