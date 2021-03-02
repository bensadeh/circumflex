package model

import (
	"clx/browser"
	"clx/cli"
	"clx/comment"
	"clx/constants/help"
	"clx/constants/messages"
	"clx/core"
	"clx/file"
	"clx/screen"
	"clx/sub"
	"clx/utils/message"
	"clx/utils/vim"
	"clx/view"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	constructor "clx/constructors"

	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

func SetAfterInitializationAndAfterResizeFunctions(app *cview.Application, list *cview.List,
	submissions []*core.Submissions, main *core.MainView, appState *core.ApplicationState, config *core.Config) {
	app.SetAfterResizeFunc(func(width int, height int) {
		if appState.IsReturningFromSuspension {
			appState.IsReturningFromSuspension = false

			return
		}
		resetStates(appState, submissions)
		initializeView(appState, submissions, main, config)
		err := fetchAndAppendSubmissionEntries(submissions[appState.SubmissionsCategory], appState, config.HideYCJobs)
		if err != nil {
			setToErrorState(appState, main, list, app)
		} else {
			appState.IsOffline = false
			showPageAfterResize(appState, list, submissions, main, config)
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

func resetStates(appState *core.ApplicationState, submissions []*core.Submissions) {
	resetApplicationState(appState)
	resetSubmissionStates(submissions)
}

func resetApplicationState(appState *core.ApplicationState) {
	appState.CurrentPage = 0
	appState.ScreenWidth = screen.GetTerminalWidth()
	appState.ScreenHeight = screen.GetTerminalHeight()
	appState.SubmissionsToShow = screen.GetSubmissionsToShow(appState.ScreenHeight, 30)
}

func resetSubmissionStates(submissions []*core.Submissions) {
	numberOfCategories := 3

	for i := 0; i < numberOfCategories; i++ {
		submissions[i].MappedSubmissions = 0
		submissions[i].PageToFetchFromAPI = 0
		submissions[i].StoriesListed = 0
		submissions[i].Entries = nil
	}
}

func initializeView(appState *core.ApplicationState, submissions []*core.Submissions, main *core.MainView,
	config *core.Config) {
	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, 0, 0)
	view.SetLeftMarginText(main, marginText)
	view.UpdateSettingsScreen(main)
	view.SetPanelToSubmissions(main)
	view.SetHackerNewsHeader(main, appState.SubmissionsCategory)
	view.SetPageCounter(main, appState.CurrentPage, submissions[appState.SubmissionsCategory].MaxPages)
}

func showPageAfterResize(appState *core.ApplicationState, list *cview.List, submissions []*core.Submissions,
	main *core.MainView, config *core.Config) {
	submissionEntries := submissions[appState.SubmissionsCategory].Entries
	statusBarText := getInfoScreenStatusBarText(appState.HelpScreenCategory)

	SetListItemsToCurrentPage(list, submissionEntries, appState.CurrentPage, appState.SubmissionsToShow, config)

	if appState.IsOnHelpScreen {
		updateInfoScreenView(main, appState.HelpScreenCategory, statusBarText)
	}
}

func ReadSubmissionComments(app *cview.Application, main *core.MainView, list *cview.List,
	submissions []*core.Submission, appState *core.ApplicationState, config *core.Config) {
	i := list.GetCurrentItemIndex()

	for index := range submissions {
		if index == i {
			storyIndex := (appState.CurrentPage)*appState.SubmissionsToShow + i
			s := submissions[storyIndex]

			if s.Author == "" {
				appState.IsReturningFromSuspension = true

				return
			}

			app.Suspend(func() {
				id := strconv.Itoa(s.ID)
				comments, err := comment.FetchComments(id)
				screenWidth := screen.GetTerminalWidth()

				if err != nil {
					errorMessage := message.Error(messages.CommentsNotFetched)
					view.SetTemporaryStatusBar(app, main, errorMessage, 4*time.Second)
				} else {
					commentTree := comment.ToString(*comments,
						config.IndentSize, config.CommentWidth, screenWidth, config.PreserveRightMargin)

					cli.Less(commentTree)
				}
			})

			appState.IsReturningFromSuspension = true
		}
	}
}

func OpenCommentsInBrowser(list *cview.List, appState *core.ApplicationState, submissions []*core.Submission) {
	item := list.GetCurrentItemIndex() + appState.SubmissionsToShow*(appState.CurrentPage)
	id := submissions[item].ID
	url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(id)
	browser.Open(url)
}

func OpenLinkInBrowser(list *cview.List, appState *core.ApplicationState, submissions []*core.Submission) {
	item := list.GetCurrentItemIndex() + appState.SubmissionsToShow*(appState.CurrentPage)
	url := submissions[item].URL
	browser.Open(url)
}

func NextPage(app *cview.Application, list *cview.List, submissions *core.Submissions, main *core.MainView,
	appState *core.ApplicationState, config *core.Config) {
	isOnLastPage := appState.CurrentPage+1 > submissions.MaxPages
	if isOnLastPage {
		return
	}

	currentlySelectedItem := list.GetCurrentItemIndex()

	if !pageHasEnoughSubmissionsToView(appState.CurrentPage+1, appState.SubmissionsToShow, submissions.Entries) {
		err := fetchAndAppendSubmissionEntries(submissions, appState, config.HideYCJobs)
		if err != nil {
			setToErrorState(appState, main, list, app)

			return
		}
	}

	appState.CurrentPage++

	SetListItemsToCurrentPage(list, submissions.Entries, appState.CurrentPage, appState.SubmissionsToShow, config)
	list.SetCurrentItem(currentlySelectedItem)

	ClearVimRegister(main, appState)

	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, currentlySelectedItem,
		appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)
	view.SetPageCounter(main, appState.CurrentPage, submissions.MaxPages)
}

func getMarginText(useRelativeNumbering bool, viewableStoriesOnSinglePage int, currentPosition int,
	currentPage int) string {
	if useRelativeNumbering {
		return vim.RelativeRankings(viewableStoriesOnSinglePage, currentPosition, currentPage)
	}

	return vim.AbsoluteRankings(viewableStoriesOnSinglePage, currentPage)
}

func pageHasEnoughSubmissionsToView(page int, visibleStories int, submissions []*core.Submission) bool {
	largestItemToDisplay := (page * visibleStories) + visibleStories
	downloadedSubmissions := len(submissions)

	return downloadedSubmissions > largestItemToDisplay
}

func fetchAndAppendSubmissionEntries(submissions *core.Submissions, appState *core.ApplicationState,
	hideYCJobs bool) error {
	submissions.PageToFetchFromAPI++

	newSubmissions, err := sub.FetchSubmissions(submissions.PageToFetchFromAPI, appState.SubmissionsCategory)
	if err != nil {
		return fmt.Errorf("could not fetch submissions: %w", err)
	}

	filteredSubmissions := sub.Filter(newSubmissions, hideYCJobs)
	submissions.Entries = append(submissions.Entries, filteredSubmissions...)

	return nil
}

func SetListItemsToCurrentPage(list *cview.List, submissions []*core.Submission, currentPage int, viewableStories int,
	config *core.Config) {
	view.ClearList(list)

	start := currentPage * viewableStories
	end := start + viewableStories

	for i := start; i < end; i++ {
		s := submissions[i]
		mainText := sub.FormatSubMain(s.Title, s.Domain, config.HighlightHeadlines)
		secondaryText := sub.FormatSubSecondary(s.Points, s.Author, s.Time, s.CommentsCount)

		item := cview.NewListItem(mainText)
		item.SetSecondaryText(secondaryText)

		list.AddItem(item)
	}
}

func ChangeCategory(app *cview.Application, event *tcell.EventKey, list *cview.List, appState *core.ApplicationState,
	submissions []*core.Submissions, main *core.MainView, config *core.Config) {
	currentItem := list.GetCurrentItemIndex()

	if event.Key() == tcell.KeyBacktab {
		appState.SubmissionsCategory = getPreviousCategory(appState.SubmissionsCategory, 4)
	} else {
		appState.SubmissionsCategory = getNextCategory(appState.SubmissionsCategory, 4)
	}

	currentSubmissions := submissions[appState.SubmissionsCategory]
	appState.CurrentPage = 0

	if !pageHasEnoughSubmissionsToView(0, appState.SubmissionsToShow, currentSubmissions.Entries) {
		err := fetchAndAppendSubmissionEntries(currentSubmissions, appState, config.HideYCJobs)
		if err != nil {
			setToErrorState(appState, main, list, app)

			return
		}
	}

	SetListItemsToCurrentPage(list, currentSubmissions.Entries, appState.CurrentPage, appState.SubmissionsToShow, config)
	list.SetCurrentItem(currentItem)
	ClearVimRegister(main, appState)

	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, currentItem, appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)

	view.SetPageCounter(main, appState.CurrentPage, currentSubmissions.MaxPages)
	view.SetHackerNewsHeader(main, appState.SubmissionsCategory)
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
		appState.HelpScreenCategory = getPreviousCategory(appState.HelpScreenCategory, 3)
	} else {
		appState.HelpScreenCategory = getNextCategory(appState.HelpScreenCategory, 3)
	}

	statusBarText := getInfoScreenStatusBarText(appState.HelpScreenCategory)

	updateInfoScreenView(main, appState.HelpScreenCategory, statusBarText)
}

func PreviousPage(list *cview.List, submissions *core.Submissions, main *core.MainView, appState *core.ApplicationState,
	config *core.Config) {
	previousPage := appState.CurrentPage - 1
	if previousPage < 0 {
		return
	}

	appState.CurrentPage--

	currentItem := list.GetCurrentItemIndex()

	SetListItemsToCurrentPage(list, submissions.Entries, appState.CurrentPage, appState.SubmissionsToShow, config)

	list.SetCurrentItem(currentItem)
	ClearVimRegister(main, appState)

	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, currentItem, appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)
	view.SetPageCounter(main, appState.CurrentPage, submissions.MaxPages)
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

	err := file.WriteToConfigFile(constructor.GetConfigFileContents())
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

	list.SetCurrentItem(nextItem)

	ClearVimRegister(main, appState)
	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, nextItem, appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)
	view.ClearStatusBar(main)
}

func SelectItemUp(main *core.MainView, list *cview.List, appState *core.ApplicationState, config *core.Config) {
	currentItem := list.GetCurrentItemIndex()
	nextItem := vim.GetItemUp(appState.VimNumberRegister, currentItem)

	list.SetCurrentItem(nextItem)

	ClearVimRegister(main, appState)
	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, nextItem, appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)
	view.ClearStatusBar(main)
}

func EnterInfoScreen(main *core.MainView, appState *core.ApplicationState) {
	statusBarText := getInfoScreenStatusBarText(appState.HelpScreenCategory)
	appState.IsOnHelpScreen = true

	ClearVimRegister(main, appState)
	updateInfoScreenView(main, appState.HelpScreenCategory, statusBarText)
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

func ExitHelpScreen(main *core.MainView, appState *core.ApplicationState, submissions *core.Submissions,
	config *core.Config, list *cview.List) {
	appState.IsOnHelpScreen = false

	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, list.GetCurrentItemIndex(),
		appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)
	view.SetHackerNewsHeader(main, appState.SubmissionsCategory)
	view.SetPanelToSubmissions(main)
	view.SetPageCounter(main, appState.CurrentPage, submissions.MaxPages)
	view.ClearStatusBar(main)
}

func SelectFirstElementInList(main *core.MainView, appState *core.ApplicationState, list *cview.List,
	config *core.Config) {
	view.SelectFirstElementInList(list)
	ClearVimRegister(main, appState)

	marginText := getMarginText(config.RelativeNumbering, appState.SubmissionsToShow, list.GetCurrentItemIndex(),
		appState.CurrentPage)
	view.SetLeftMarginText(main, marginText)
}

func GoToLowerCaseG(main *core.MainView, appState *core.ApplicationState, list *cview.List, config *core.Config) {
	switch {
	case appState.VimNumberRegister == "g":
		SelectFirstElementInList(main, appState, list, config)

		marginText := getMarginText(config.RelativeNumbering,
			appState.SubmissionsToShow,
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
		list.SetCurrentItem(itemToJumpTo)
		view.ClearStatusBar(main)

		marginText := getMarginText(config.RelativeNumbering,
			appState.SubmissionsToShow,
			list.GetCurrentItemIndex(),
			appState.CurrentPage)
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

		marginText := getMarginText(config.RelativeNumbering,
			appState.SubmissionsToShow,
			list.GetCurrentItemIndex(),
			appState.CurrentPage)
		view.SetLeftMarginText(main, marginText)

	case vim.ContainsOnlyNumbers(appState.VimNumberRegister):
		register := strings.TrimSuffix(appState.VimNumberRegister, "g")

		itemToJumpTo := vim.GetItemToJumpTo(register,
			list.GetCurrentItemIndex(),
			appState.SubmissionsToShow,
			appState.CurrentPage)

		ClearVimRegister(main, appState)
		list.SetCurrentItem(itemToJumpTo)
		view.ClearStatusBar(main)

		marginText := getMarginText(config.RelativeNumbering,
			appState.SubmissionsToShow,
			list.GetCurrentItemIndex(),
			appState.CurrentPage)
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

func Refresh(app *cview.Application, list *cview.List, main *core.MainView, submissions []*core.Submissions,
	appState *core.ApplicationState, config *core.Config) {
	afterResizeFunc := app.GetAfterResizeFunc()
	afterResizeFunc(appState.ScreenWidth, appState.ScreenHeight)

	ExitHelpScreen(main, appState, submissions[appState.SubmissionsCategory], config, list)

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
