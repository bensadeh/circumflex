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
		isOnSettingsPage := isOnHelpScreen && (appState.CurrentHelpScreenCategory == categories.Settings)

		switch {
		// Offline
		case appState.State == state.Offline && event.Rune() == 'r':
			model.Refresh(app, list, main, appState, config, ret)

		case appState.State == state.Offline && event.Rune() == 'q':
			model.Quit(app)
		case appState.State == state.Offline:
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

		case isOnHelpScreen && (event.Rune() == 'i' || event.Key() == tcell.KeyEsc || event.Rune() == '?'):
			model.ExitInfoScreen(main, appState, config, list, ret)

		case isOnHelpScreen && (event.Rune() == 'q'):
			model.Quit(app)

		case isOnHelpScreen:
			return event

		// Main View
		case appState.IsOnAddFavoriteByID:
			return event

		case appState.IsOnAddFavoriteConfirmationMessage && event.Rune() == 'y':
			model.AddToFavorites(app, list, main, appState, config, ret, reg)

		case appState.IsOnDeleteFavoriteConfirmationMessage && event.Rune() == 'y':
			model.DeleteItem(app, list, appState, main, config, ret, reg)

		case appState.IsOnAddFavoriteConfirmationMessage || appState.IsOnDeleteFavoriteConfirmationMessage:
			model.CancelConfirmation(appState, main)

		case event.Rune() == 'f':
			model.AddToFavoritesConfirmationDialogue(main, appState)

		case event.Rune() == 'F':
			model.ShowAddCustomFavorite(app, list, main, appState, config, ret, reg)

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
			model.Refresh(app, list, main, appState, config, ret)

		case event.Key() == tcell.KeyEnter:
			model.ReadSubmissionComments(app, main, list, appState, config, ret, reg)

		case event.Rune() == ' ':
			model.ReadSubmissionContent(app, main, list, appState, config, ret, reg)

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
