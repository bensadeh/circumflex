package controller

import (
	"clx/constants/help"
	"clx/core"
	"clx/favorites"
	"clx/model"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

func SetAfterInitializationAndAfterResizeFunctions(fav *favorites.Favorites, app *cview.Application, list *cview.List,
	submissions []*core.Submissions, main *core.MainView, appState *core.ApplicationState, config *core.Config) {
	model.SetAfterInitializationAndAfterResizeFunctions(app, list, submissions, main, appState, config)
}

func SetApplicationShortcuts(fav *favorites.Favorites, app *cview.Application, list *cview.List, submissions []*core.Submissions,
	main *core.MainView, appState *core.ApplicationState, config *core.Config) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentState := submissions[appState.SubmissionsCategory]
		isOnHelpScreen := appState.IsOnHelpScreen
		isOnSettingsPage := isOnHelpScreen && (appState.HelpScreenCategory == help.Settings)

		switch {
		// Offline
		case appState.IsOffline && event.Rune() == 'r':
			model.Refresh(app, list, main, submissions, appState, config)

		case appState.IsOffline && event.Rune() == 'q':
			model.Quit(app)
		case appState.IsOffline:
			return event

		// Help screen
		case appState.IsOnConfigCreationConfirmationMessage && event.Rune() == 'y':
			model.CreateConfig(appState, main)

		case appState.IsOnConfigCreationConfirmationMessage:
			model.CancelCreateConfigConfirmationMessage(appState, main)

		case isOnSettingsPage && event.Rune() == 't':
			model.ShowCreateConfigConfirmationMessage(main, appState)

		case isOnSettingsPage && (event.Rune() == 'j' || event.Key() == tcell.KeyDown):
			model.ScrollSettingsOneLineDown(main)

		case isOnSettingsPage && (event.Rune() == 'k' || event.Key() == tcell.KeyUp):
			model.ScrollSettingsOneLineUp(main)

		case isOnSettingsPage && event.Rune() == 'd':
			model.ScrollSettingsOneHalfPageDown(main)

		case isOnSettingsPage && event.Rune() == 'u':
			model.ScrollSettingsOneHalfPageUp(main)

		case isOnSettingsPage && event.Rune() == 'g':
			model.ScrollSettingsToBeginning(main)

		case isOnSettingsPage && event.Rune() == 'G':
			model.ScrollSettingsToEnd(main)

		case isOnHelpScreen && (event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab):
			model.ChangeHelpScreenCategory(event, appState, main)

		case isOnHelpScreen && (event.Rune() == 'i'):
			model.ExitHelpScreen(main, appState, currentState, config, list)

		case isOnHelpScreen && (event.Rune() == 'q'):
			model.Quit(app)

		case isOnHelpScreen:
			return event

		// Submissions
		case event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab:
			model.ChangeCategory(app, event, list, appState, submissions, main, config)

		case event.Rune() == 'l' || event.Key() == tcell.KeyRight:
			model.NextPage(app, list, currentState, main, appState, config)

		case event.Rune() == 'h' || event.Key() == tcell.KeyLeft:
			model.PreviousPage(list, currentState, main, appState, config)

		case event.Rune() == 'j' || event.Key() == tcell.KeyDown:
			model.SelectItemDown(main, list, appState, config)

		case event.Rune() == 'k' || event.Key() == tcell.KeyUp:
			model.SelectItemUp(main, list, appState, config)

		case event.Rune() == 'q':
			model.Quit(app)

		case event.Key() == tcell.KeyEsc:
			model.ClearVimRegister(main, appState)

		case event.Rune() == 'i' || event.Rune() == '?':
			model.EnterInfoScreen(main, appState)

		case event.Rune() == 'g':
			model.GoToLowerCaseG(main, appState, list, config)

		case event.Rune() == 'G':
			model.GoToUpperCaseG(main, appState, list, config)

		case event.Rune() == 'r':
			model.Refresh(app, list, main, submissions, appState, config)

		case event.Key() == tcell.KeyEnter:
			model.ReadSubmissionComments(app, main, list, currentState.Entries, appState, config)

		case event.Rune() == 'o':
			model.OpenLinkInBrowser(list, appState, currentState.Entries)

		case event.Rune() == 'c':
			model.OpenCommentsInBrowser(list, appState, currentState.Entries)

		case unicode.IsDigit(event.Rune()):
			model.PutDigitInRegister(main, event.Rune(), appState)
		}

		return event
	})
}
