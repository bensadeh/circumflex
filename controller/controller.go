package controller

import (
	"clx/constants/categories"
	"clx/constants/state"
	"clx/core"
	"clx/handler"
	"clx/model"
	"clx/utils/vim"
	"unicode"

	"code.rocketnine.space/tslocum/cview"
	"github.com/gdamore/tcell/v2"
)

func SetAfterInitializationAndAfterResizeFunctions(ret *handler.StoryHandler,
	app *cview.Application, list *cview.List, main *core.MainView, appState *core.ApplicationState, config *core.Config) {
	model.SetAfterInitializationAndAfterResizeFunctions(app, list, main, appState, config, ret)
}

func SetApplicationShortcuts(ret *handler.StoryHandler, reg *vim.Register, app *cview.Application, list *cview.List,
	main *core.MainView, appState *core.ApplicationState, config *core.Config) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		isOnHelpScreen := appState.State == state.OnHelpScreen

		switch {
		// Offline
		case appState.IsOffline && event.Rune() == 'r':
			model.Refresh(app, main, appState)

		case appState.IsOffline && event.Rune() == 'q':
			model.Quit(app)

		case appState.IsOffline:
			return event

		// Help View
		case isOnHelpScreen && (event.Rune() == 'j' || event.Key() == tcell.KeyDown):
			model.ScrollSettingsOneLineDown(main)

		case isOnHelpScreen && (event.Rune() == 'k' || event.Key() == tcell.KeyUp):
			model.ScrollSettingsOneLineUp(main)

		case isOnHelpScreen && event.Rune() == 'd':
			model.ScrollSettingsOneHalfPageDown(main)

		case isOnHelpScreen && event.Rune() == 'u':
			model.ScrollSettingsOneHalfPageUp(main)

		case isOnHelpScreen && event.Rune() == 'g':
			model.ScrollSettingsToBeginning(main)

		case isOnHelpScreen && event.Rune() == 'G':
			model.ScrollSettingsToEnd(main)

		case isOnHelpScreen && (event.Rune() == 'i' || event.Key() == tcell.KeyEsc ||
			event.Rune() == '?' || event.Rune() == 'q'):
			model.ExitInfoScreen(main, appState, config, list, ret)

		case isOnHelpScreen:
			return event

		// Main View
		case appState.IsOnAddFavoriteConfirmationMessage && event.Rune() == 'y':
			model.AddToFavorites(app, list, main, appState, config, ret, reg)

		case appState.IsOnDeleteFavoriteConfirmationMessage && event.Rune() == 'y':
			model.DeleteItem(app, list, appState, main, config, ret, reg)

		case appState.IsOnAddFavoriteConfirmationMessage || appState.IsOnDeleteFavoriteConfirmationMessage:
			model.CancelConfirmation(appState, main)

		case event.Rune() == 'f':
			model.AddToFavoritesConfirmationDialogue(main, appState)

		case event.Rune() == 'x' && appState.CurrentCategory == categories.Favorites:
			model.DeleteFavoriteConfirmationDialogue(main, appState)

		case event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab:
			model.ChangeCategory(app, event, list, appState, main, config, ret, reg)

		case event.Rune() == 'l' || event.Key() == tcell.KeyRight:
			model.NextPage(app, list, main, appState, config, ret, reg)

		case event.Rune() == 'h' || event.Key() == tcell.KeyLeft:
			model.PreviousPage(app, list, main, appState, config, ret, reg)

		case event.Rune() == 'k' || event.Key() == tcell.KeyUp:
			model.SelectItemUp(main, list, appState, config, reg)

		case event.Rune() == 'j' || event.Key() == tcell.KeyDown:
			model.SelectItemDown(main, list, appState, config, reg)

		case event.Rune() == 'q':
			model.Quit(app)

		case event.Key() == tcell.KeyEsc:
			model.ClearVimRegister(main, reg)

		case event.Rune() == 'i' || event.Rune() == '?':
			model.EnterInfoScreen(main, appState, reg)

		case event.Rune() == 'g':
			model.LowerCaseG(main, appState, list, config, reg)

		case event.Rune() == 'G':
			model.UpperCaseG(main, appState, list, config, reg, ret)

		case event.Rune() == 'r':
			model.Refresh(app, main, appState)

		case event.Key() == tcell.KeyEnter:
			model.ReadSubmissionComments(app, main, list, appState, config, ret, reg)

		case event.Rune() == ' ':
			model.ReadSubmissionContent(app, main, list, appState, config, ret, reg)

		case event.Rune() == 't':
			model.ForceReadSubmissionContent(app, main, list, appState, config, ret, reg)

		case event.Rune() == 'u':
			model.ForceReadSubmissionContentNew(app, main, list, appState, config, ret, reg)

		case event.Rune() == 'o':
			model.OpenLinkInBrowser(list, appState, ret)

		case event.Rune() == 'c':
			model.OpenCommentsInBrowser(list, appState, ret)

		case unicode.IsDigit(event.Rune()):
			model.PutDigitInRegister(main, event.Rune(), reg)
		}

		return event
	})
}
