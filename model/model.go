package model

import (
	"clx/browser"
	"clx/cli"
	"clx/comment"
	"clx/constants/help"
	"clx/constants/messages"
	"clx/core"
	"clx/file"
	"clx/retriever"
	"clx/screen"
	"clx/utils/message"
	"clx/utils/vim"
	"clx/view"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	constructor "clx/constructors"

	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

func SetAfterInitializationAndAfterResizeFunctions(app *cview.Application, list *cview.List,
	main *core.MainView, appState *core.ApplicationState, config *core.Config,
	ret *retriever.Retriever) {
	app.SetAfterResizeFunc(func(width int, height int) {
		if appState.IsReturningFromSuspension {
			appState.IsReturningFromSuspension = false

			return
		}

		resetStates(appState, ret)
		initializeView(appState, main, ret)

		listItems, err := ret.GetSubmissions(appState.CurrentCategory, appState.CurrentPage,
			appState.SubmissionsToShow, config.HighlightHeadlines, config.HideYCJobs)
		if err != nil {
			setToErrorState(appState, main, list, app)

			return
		}

		appState.IsOffline = false
		statusBarText := getInfoScreenStatusBarText(appState.CurrentHelpScreenCategory)
		marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, len(listItems), 0, 0)

		view.ShowItems(list, listItems)
		view.SetLeftMarginText(main, marginText)

		if appState.IsOnHelpScreen {
			updateInfoScreenView(main, appState.CurrentHelpScreenCategory, statusBarText)
		}
	})
}

func setToErrorState(appState *core.ApplicationState, main *core.MainView, list *cview.List, app *cview.Application) {
	errorMessage := message.Error(messages.OfflineMessage)
	appState.IsOffline = true

	view.SetPermanentStatusBar(main, errorMessage, cview.AlignCenter)
	view.ClearList(list)
	app.Draw()
}

func resetStates(appState *core.ApplicationState, ret *retriever.Retriever) {
	resetApplicationState(appState)
	ret.Reset()
}

func resetApplicationState(appState *core.ApplicationState) {
	appState.CurrentPage = 0
	appState.ScreenWidth = screen.GetTerminalWidth()
	appState.ScreenHeight = screen.GetTerminalHeight()
	appState.SubmissionsToShow = screen.GetSubmissionsToShow(appState.ScreenHeight, 30)
}

func initializeView(appState *core.ApplicationState, main *core.MainView, ret *retriever.Retriever) {
	view.UpdateSettingsScreen(main)
	view.UpdateInfoScreen(main)
	view.SetPanelToSubmissions(main)
	view.SetHackerNewsHeader(main, appState.CurrentCategory)
	view.SetPageCounter(main, appState.CurrentPage, ret.GetMaxPages(appState.CurrentCategory,
		appState.SubmissionsToShow))
}

func ReadSubmissionComments(app *cview.Application, main *core.MainView, list *cview.List,
	appState *core.ApplicationState, config *core.Config, r *retriever.Retriever) {
	story := r.GetStory(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.SubmissionsToShow,
		appState.CurrentPage)

	if story.Author == "" {
		appState.IsReturningFromSuspension = true

		return
	}

	app.Suspend(func() {
		id := strconv.Itoa(story.ID)
		screenWidth := screen.GetTerminalWidth()

		comments, err := comment.FetchComments(id)
		if err != nil {
			errorMessage := message.Error(messages.CommentsNotFetched)
			view.SetTemporaryStatusBar(app, main, errorMessage, 4*time.Second)

			return
		}

		r.UpdateFavoriteStoryAndWriteToDisk(comments)
		commentTree := comment.ToString(*comments, config.IndentSize, config.CommentWidth, screenWidth,
			config.PreserveRightMargin)

		cli.Less(commentTree)
	})

	appState.IsReturningFromSuspension = true
}

func OpenCommentsInBrowser(list *cview.List, appState *core.ApplicationState, r *retriever.Retriever) {
	story := r.GetStory(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.SubmissionsToShow,
		appState.CurrentPage)
	url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(story.ID)
	browser.Open(url)
}

func OpenLinkInBrowser(list *cview.List, appState *core.ApplicationState, r *retriever.Retriever) {
	story := r.GetStory(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.SubmissionsToShow,
		appState.CurrentPage)
	browser.Open(story.URL)
}

func NextPage(app *cview.Application, list *cview.List, main *core.MainView, appState *core.ApplicationState,
	config *core.Config, ret *retriever.Retriever) {
	isOnLastPage := appState.CurrentPage+1 > ret.GetMaxPages(appState.CurrentCategory, appState.SubmissionsToShow)
	if isOnLastPage {
		return
	}

	changePage(app, list, main, appState, config, ret, 1)
}

func PreviousPage(app *cview.Application, list *cview.List, main *core.MainView, appState *core.ApplicationState,
	config *core.Config, ret *retriever.Retriever) {
	isOnFirstPage := appState.CurrentPage-1 < 0
	if isOnFirstPage {
		return
	}

	changePage(app, list, main, appState, config, ret, -1)
}

func changePage(app *cview.Application, list *cview.List, main *core.MainView, appState *core.ApplicationState,
	config *core.Config, ret *retriever.Retriever, delta int) {
	currentlySelectedItem := list.GetCurrentItemIndex()
	appState.CurrentPage += delta

	listItems, err := ret.GetSubmissions(appState.CurrentCategory, appState.CurrentPage,
		appState.SubmissionsToShow, config.HighlightHeadlines, config.HideYCJobs)
	if err != nil {
		setToErrorState(appState, main, list, app)

		return
	}

	view.ShowItems(list, listItems)
	view.SelectItem(list, currentlySelectedItem)

	ClearVimRegister(main, appState)

	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, len(listItems), currentlySelectedItem,
		appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)
	view.SetPageCounter(main, appState.CurrentPage, ret.GetMaxPages(appState.CurrentCategory,
		appState.SubmissionsToShow))
}

func getMarginText(useRelativeNumbering bool, viewableStories, maxItems, currentPosition, currentPage int) string {
	if useRelativeNumbering {
		return vim.RelativeRankings(maxItems, currentPosition, currentPage)
	}

	return vim.AbsoluteRankings(viewableStories, maxItems, currentPage)
}

func ChangeCategory(app *cview.Application, event *tcell.EventKey, list *cview.List, appState *core.ApplicationState,
	main *core.MainView, config *core.Config, ret *retriever.Retriever) {
	currentItem := list.GetCurrentItemIndex()
	nextCategory := 0

	if event.Key() == tcell.KeyBacktab {
		nextCategory = getPreviousCategory(appState.CurrentCategory, 5)
	} else {
		nextCategory = getNextCategory(appState.CurrentCategory, 5)
	}

	appState.CurrentCategory = nextCategory
	appState.CurrentPage = 0

	listItems, err := ret.GetSubmissions(appState.CurrentCategory, appState.CurrentPage,
		appState.SubmissionsToShow, config.HighlightHeadlines, config.HideYCJobs)
	if err != nil {
		setToErrorState(appState, main, list, app)

		return
	}

	view.ShowItems(list, listItems)
	view.SelectItem(list, currentItem)
	ClearVimRegister(main, appState)

	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, len(listItems), currentItem,
		appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)

	view.SetPageCounter(main, appState.CurrentPage, ret.GetMaxPages(appState.CurrentCategory, appState.SubmissionsToShow))
	view.SetHackerNewsHeader(main, appState.CurrentCategory)
}

func getNextCategory(currentCategory int, numberOfCategories int) int {
	if currentCategory == (numberOfCategories - 1) {
		return 0
	}

	return currentCategory + 1
}

func getPreviousCategory(currentCategory int, numberOfCategories int) int {
	if currentCategory == 0 {
		return numberOfCategories - 1
	}

	return currentCategory - 1
}

func ChangeHelpScreenCategory(event *tcell.EventKey, appState *core.ApplicationState, main *core.MainView) {
	if event.Key() == tcell.KeyBacktab {
		appState.CurrentHelpScreenCategory = getPreviousCategory(appState.CurrentHelpScreenCategory, 3)
	} else {
		appState.CurrentHelpScreenCategory = getNextCategory(appState.CurrentHelpScreenCategory, 3)
	}

	statusBarText := getInfoScreenStatusBarText(appState.CurrentHelpScreenCategory)

	updateInfoScreenView(main, appState.CurrentHelpScreenCategory, statusBarText)
}

func ShowCreateConfigConfirmationMessage(main *core.MainView, appState *core.ApplicationState) {
	if file.ConfigFileExists() {
		return
	}

	appState.IsOnConfigCreationConfirmationMessage = true

	view.SetPermanentStatusBar(main,
		"[::b]config.env[::-] will be created in [::r]~/.config/circumflex[::-], press Y to Confirm", cview.AlignCenter)
}

func ScrollSettingsOneLineUp(main *core.MainView) {
	view.ScrollSettingsOneLineUp(main)
}

func ScrollSettingsOneLineDown(main *core.MainView) {
	view.ScrollSettingsOneLineDown(main)
}

func ScrollSettingsOneHalfPageUp(main *core.MainView) {
	halfPage := screen.GetTerminalHeight() / 2
	view.ScrollSettingsByAmount(main, -halfPage)
}

func ScrollSettingsOneHalfPageDown(main *core.MainView) {
	halfPage := screen.GetTerminalHeight() / 2
	view.ScrollSettingsByAmount(main, halfPage)
}

func ScrollSettingsToBeginning(main *core.MainView) {
	view.ScrollSettingsToBeginning(main)
}

func ScrollSettingsToEnd(main *core.MainView) {
	view.ScrollSettingsToEnd(main)
}

func CancelCreateConfigConfirmationMessage(appState *core.ApplicationState, main *core.MainView) {
	appState.IsOnConfigCreationConfirmationMessage = false

	view.SetPermanentStatusBar(main, "", cview.AlignCenter)
}

func CreateConfig(appState *core.ApplicationState, main *core.MainView) {
	statusBarMessage := ""
	appState.IsOnConfigCreationConfirmationMessage = false

	err := file.WriteToFile(file.PathToConfigFile(), constructor.GetConfigFileContents())
	if err != nil {
		statusBarMessage = message.Error(messages.ConfigNotCreated)
	} else {
		statusBarMessage = message.Success(messages.ConfigCreatedAt)
	}

	view.UpdateSettingsScreen(main)
	view.SetPermanentStatusBar(main, statusBarMessage, cview.AlignCenter)
}

func SelectItemDown(main *core.MainView, list *cview.List, appState *core.ApplicationState, config *core.Config) {
	currentItem := list.GetCurrentItemIndex()
	itemCount := list.GetItemCount()
	nextItem := vim.GetItemDown(appState.VimNumberRegister, currentItem, itemCount)

	view.SelectItem(list, nextItem)

	ClearVimRegister(main, appState)
	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, list.GetItemCount(), nextItem,
		appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)
	view.ClearStatusBar(main)
}

func SelectItemUp(main *core.MainView, list *cview.List, appState *core.ApplicationState, config *core.Config) {
	currentItem := list.GetCurrentItemIndex()
	nextItem := vim.GetItemUp(appState.VimNumberRegister, currentItem)

	view.SelectItem(list, nextItem)

	ClearVimRegister(main, appState)
	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, list.GetItemCount(), nextItem,
		appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)
	view.ClearStatusBar(main)
}

func EnterInfoScreen(main *core.MainView, appState *core.ApplicationState) {
	statusBarText := getInfoScreenStatusBarText(appState.CurrentHelpScreenCategory)
	appState.IsOnHelpScreen = true

	ClearVimRegister(main, appState)
	updateInfoScreenView(main, appState.CurrentHelpScreenCategory, statusBarText)
}

func getInfoScreenStatusBarText(category int) string {
	if category == help.Info {
		return messages.GetCircumflexStatusMessage()
	}

	return ""
}

func updateInfoScreenView(main *core.MainView, helpScreenCategory int, statusBarText string) {
	view.SetPermanentStatusBar(main, statusBarText, cview.AlignCenter)
	view.HidePageCounter(main)
	view.SetHelpScreenHeader(main, helpScreenCategory)
	view.HideLeftMarginRanks(main)
	view.SetHelpScreenPanel(main, helpScreenCategory)
}

func ExitHelpScreen(main *core.MainView, appState *core.ApplicationState, config *core.Config, list *cview.List,
	ret *retriever.Retriever) {
	appState.IsOnHelpScreen = false

	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, list.GetItemCount(),
		list.GetCurrentItemIndex(), appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)
	view.SetHackerNewsHeader(main, appState.CurrentCategory)
	view.SetPanelToSubmissions(main)
	view.SetPageCounter(main, appState.CurrentPage, ret.GetMaxPages(appState.CurrentCategory,
		appState.SubmissionsToShow))
	view.ClearStatusBar(main)
}

func SelectFirstElementInList(main *core.MainView, appState *core.ApplicationState, list *cview.List,
	config *core.Config) {
	view.SelectFirstElementInList(list)
	ClearVimRegister(main, appState)

	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, list.GetItemCount(),
		list.GetCurrentItemIndex(), appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)
}

func GoToLowerCaseG(main *core.MainView, appState *core.ApplicationState, list *cview.List, config *core.Config) {
	switch {
	case appState.VimNumberRegister == "g":
		SelectFirstElementInList(main, appState, list, config)

		marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, list.GetItemCount(),
			list.GetCurrentItemIndex(),
			appState.CurrentPage)

		view.SetLeftMarginText(main, marginText)
		view.ClearStatusBar(main)

	case vim.ContainsOnlyNumbers(appState.VimNumberRegister):
		appState.VimNumberRegister += "g"

		view.SetPermanentStatusBar(main, vim.FormatRegisterOutput(appState.VimNumberRegister), cview.AlignRight)

	case vim.IsNumberWithGAppended(appState.VimNumberRegister):
		register := strings.TrimSuffix(appState.VimNumberRegister, "g")

		itemToJumpTo := vim.GetItemToJumpTo(register,
			list.GetCurrentItemIndex(),
			appState.SubmissionsToShow,
			appState.CurrentPage)

		ClearVimRegister(main, appState)
		view.SelectItem(list, itemToJumpTo)
		view.ClearStatusBar(main)

		marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, list.GetItemCount(),
			list.GetCurrentItemIndex(), appState.CurrentPage)
		view.SetLeftMarginText(main, marginText)

	case appState.VimNumberRegister == "":
		appState.VimNumberRegister += "g"

		view.SetPermanentStatusBar(main, vim.FormatRegisterOutput(appState.VimNumberRegister), cview.AlignRight)
	}
}

func GoToUpperCaseG(main *core.MainView, appState *core.ApplicationState, list *cview.List, config *core.Config) {
	switch {
	case appState.VimNumberRegister == "":
		view.SelectLastElementInList(list)
		ClearVimRegister(main, appState)

		marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, list.GetItemCount(),
			list.GetCurrentItemIndex(), appState.CurrentPage)
		view.SetLeftMarginText(main, marginText)

	case vim.ContainsOnlyNumbers(appState.VimNumberRegister):
		register := strings.TrimSuffix(appState.VimNumberRegister, "g")

		itemToJumpTo := vim.GetItemToJumpTo(register, list.GetCurrentItemIndex(), appState.SubmissionsToShow,
			appState.CurrentPage)

		ClearVimRegister(main, appState)
		view.SelectItem(list, itemToJumpTo)
		view.ClearStatusBar(main)

		marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, list.GetItemCount(),
			list.GetCurrentItemIndex(), appState.CurrentPage)
		view.SetLeftMarginText(main, marginText)
	case vim.IsNumberWithGAppended(appState.VimNumberRegister):
		ClearVimRegister(main, appState)
		view.ClearStatusBar(main)

	case appState.VimNumberRegister == "g":
		ClearVimRegister(main, appState)
		view.ClearStatusBar(main)
	}
}

func PutDigitInRegister(main *core.MainView, element rune, appState *core.ApplicationState) {
	if len(appState.VimNumberRegister) == 0 && string(element) == "0" {
		return
	}

	if appState.VimNumberRegister == "g" {
		ClearVimRegister(main, appState)
	}

	registerIsMoreThanThreeDigits := len(appState.VimNumberRegister) > 2

	if registerIsMoreThanThreeDigits {
		appState.VimNumberRegister = trimFirstRune(appState.VimNumberRegister)
	}

	appState.VimNumberRegister += string(element)

	view.SetPermanentStatusBar(main, vim.FormatRegisterOutput(appState.VimNumberRegister), cview.AlignRight)
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)

	return s[i:]
}

func Quit(app *cview.Application) {
	app.Stop()
}

func ClearVimRegister(main *core.MainView, appState *core.ApplicationState) {
	appState.VimNumberRegister = ""

	view.ClearStatusBar(main)
}

func Refresh(app *cview.Application, list *cview.List, main *core.MainView, appState *core.ApplicationState,
	config *core.Config, ret *retriever.Retriever) {
	afterResizeFunc := app.GetAfterResizeFunc()
	afterResizeFunc(appState.ScreenWidth, appState.ScreenHeight)

	ExitHelpScreen(main, appState, config, list, ret)

	if appState.IsOffline {
		errorMessage := message.Error(messages.OfflineMessage)

		view.SetPermanentStatusBar(main, errorMessage, cview.AlignCenter)
		view.ClearList(list)
		app.Draw()
	} else {
		duration := time.Millisecond * 2000
		view.SetTemporaryStatusBar(app, main, "Refreshed", duration)
	}
}

func AddToFavoritesConfirmationDialogue(main *core.MainView, appState *core.ApplicationState) {
	appState.IsOnFavoritesConfirmationMessage = true

	view.SetPermanentStatusBar(main,
		"Highlighted item will be added to Favorites, press Y to Confirm", cview.AlignCenter)
}

func AddToFavorites(list *cview.List, main *core.MainView, appState *core.ApplicationState, ret *retriever.Retriever) {
	statusBarMessage := ""
	appState.IsOnFavoritesConfirmationMessage = false

	story := ret.GetStory(appState.CurrentCategory, list.GetCurrentItemIndex(), appState.SubmissionsToShow,
		appState.CurrentPage)
	ret.AddItemToFavorites(story)
	bytes, _ := ret.GetFavoritesJSON()
	filePath := file.PathToFavoritesFile()

	err := file.WriteToFile(filePath, string(bytes))
	if err != nil {
		statusBarMessage = message.Error("Could not add to favorites")
	} else {
		statusBarMessage = message.Success("Item added to favorites")
	}

	view.UpdateSettingsScreen(main)
	view.SetPermanentStatusBar(main, statusBarMessage, cview.AlignCenter)
}
