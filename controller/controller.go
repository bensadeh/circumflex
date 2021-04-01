package controller

import (
	"clx/constants/categories"
	"clx/constants/help"
	"clx/core"
	"clx/model"
	"clx/retriever"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

func SetAfterInitializationAndAfterResizeFunctions(ret *retriever.Retriever,
	app *cview.Application, list *cview.List, main *core.MainView, appState *core.ApplicationState, config *core.Config) {
	model.SetAfterInitializationAndAfterResizeFunctions(app, list, main, appState, config, ret)
}

func SetApplicationShortcuts(ret *retriever.Retriever, app *cview.Application, list *cview.List,
	main *core.MainView, appState *core.ApplicationState, config *core.Config) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		isOnHelpScreen := appState.IsOnHelpScreen
		isOnSettingsPage := isOnHelpScreen && (appState.CurrentHelpScreenCategory == help.Settings)

		switch {
		// Offline
		case appState.IsOffline && event.Rune() == 'r':
			model.Refresh(app, list, main, appState, config, ret)

		case appState.IsOffline && event.Rune() == 'q':
			model.Quit(app)
		case appState.IsOffline:
			return event

		// Help View
		case appState.IsOnConfigCreationConfirmationMessage && event.Rune() == 'y':
			model.CreateConfig(appState, main)

		case appState.IsOnConfigCreationConfirmationMessage:
			model.CancelConfirmation(appState, main)

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
			model.ExitHelpScreen(main, appState, config, list, ret)

		case isOnHelpScreen && (event.Rune() == 'q'):
			model.Quit(app)

		case isOnHelpScreen:
			return event

		// Main View
		case appState.IsOnAddFavoriteConfirmationMessage && event.Rune() == 'y':
			model.AddToFavorites(app, list, main, appState, config, ret)

		case appState.IsOnDeleteFavoriteConfirmationMessage && event.Rune() == 'y':
			model.DeleteItem(app, list, appState, main, config, ret)

		case appState.IsOnAddFavoriteConfirmationMessage || appState.IsOnDeleteFavoriteConfirmationMessage:
			model.CancelConfirmation(appState, main)

		case event.Rune() == 'f':
			model.AddToFavoritesConfirmationDialogue(main, appState, list)

		case event.Rune() == 'x' && appState.CurrentCategory == categories.Favorites:
			model.DeleteFavoriteConfirmationDialogue(main, appState, list)

		case event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab:
			model.ChangeCategory(app, event, list, appState, main, config, ret)

		case event.Rune() == 'l' || event.Key() == tcell.KeyRight:
			model.NextPage(app, list, main, appState, config, ret)

		case event.Rune() == 'h' || event.Key() == tcell.KeyLeft:
			model.PreviousPage(app, list, main, appState, config, ret)

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
			model.Refresh(app, list, main, appState, config, ret)

		case event.Key() == tcell.KeyEnter:
			model.ReadSubmissionComments(app, main, list, appState, config, ret)

		case event.Rune() == 'o':
			model.OpenLinkInBrowser(list, appState, ret)

		case event.Rune() == 'c':
			model.OpenCommentsInBrowser(list, appState, ret)

		case unicode.IsDigit(event.Rune()):
			model.PutDigitInRegister(main, event.Rune(), appState)
		}

		return event
	})
}
