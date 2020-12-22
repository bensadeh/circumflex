package model

import (
	"clx/browser"
	"clx/cli"
	cp "clx/comment-parser"
	"clx/constants"
	"clx/http"
	"clx/screen"
	"clx/settings"
	"clx/structs"
	"clx/submission/fetcher"
	"clx/submission/formatter"
	"clx/view"
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"strconv"
	"time"
)

func SetAfterInitializationAndAfterResizeFunctions(
	app *cview.Application,
	list *cview.List,
	submissions []*structs.Submissions,
	main *structs.MainView,
	appState *structs.ApplicationState) {
	app.SetAfterResizeFunc(func(width int, height int) {
		if appState.IsReturningFromSuspension {
			appState.IsReturningFromSuspension = false
			return
		}
		resetStates(appState, submissions)
		initializeHeaderAndFooterAndLeftMarginView(appState, submissions, main)
		err := fetchAndAppendSubmissionEntries(submissions[appState.SubmissionsCategory], appState)
		if err != nil {
			setApplicationToErrorState(appState, main, list, app)
		} else {
			appState.IsOffline = false
			showPageAfterResize(appState, list, submissions, main)
		}
	})
}

func setApplicationToErrorState(
	appState *structs.ApplicationState,
	main *structs.MainView,
	list *cview.List,
	app *cview.Application) {

	appState.IsOffline = true
	list.Clear()
	view.SetPermanentStatusBar(main, constants.OfflineMessage)
	app.Draw()
}

func resetStates(appState *structs.ApplicationState, submissions []*structs.Submissions) {
	resetApplicationState(appState)
	resetSubmissionStates(submissions)
}

func resetApplicationState(appState *structs.ApplicationState) {
	appState.CurrentPage = 0
	appState.ScreenWidth = screen.GetTerminalWidth()
	appState.ScreenHeight = screen.GetTerminalHeight()
	appState.SubmissionsToShow = screen.GetSubmissionsToShow(appState.ScreenHeight, 30)
}

func resetSubmissionStates(submissions []*structs.Submissions) {
	numberOfCategories := 3

	for i := 0; i < numberOfCategories; i++ {
		submissions[i].MappedSubmissions = 0
		submissions[i].PageToFetchFromAPI = 0
		submissions[i].StoriesListed = 0
		submissions[i].Entries = nil
	}
}

func initializeHeaderAndFooterAndLeftMarginView(
	appState *structs.ApplicationState,
	submissions []*structs.Submissions,
	main *structs.MainView) {
	view.SetPanelToSubmissions(main)
	view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.SubmissionsCategory)
	view.SetLeftMarginRanks(main, 0, appState.SubmissionsToShow)
	view.SetPageCounter(main, 0, submissions[appState.SubmissionsCategory].MaxPages, "orange")
}

func showPageAfterResize(
	appState *structs.ApplicationState,
	list *cview.List,
	submissions []*structs.Submissions,
	main *structs.MainView) {
	submissionEntries := submissions[appState.SubmissionsCategory].Entries

	SetListItemsToCurrentPage(list, submissionEntries, appState.CurrentPage, appState.SubmissionsToShow)

	if appState.IsOnHelpScreen {
		showInfoCategory(main, appState)
	}
}

func ReadSubmissionComments(
	app *cview.Application,
	list *cview.List,
	submissions []*structs.Submission,
	appState *structs.ApplicationState) {
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
				JSON, _ := http.Get("http://api.hackerwebapp.com/item/" + id)
				jComments := new(cp.Comments)
				_ = json.Unmarshal(JSON, jComments)

				commentTree := cp.PrintCommentTree(*jComments, 4, 70)

				cli.Less(commentTree)
			})

			appState.IsReturningFromSuspension = true
		}
	}
}

func OpenCommentsInBrowser(list *cview.List, appState *structs.ApplicationState, submissions []*structs.Submission) {
	item := list.GetCurrentItemIndex() + appState.SubmissionsToShow*(appState.CurrentPage)
	id := submissions[item].ID
	url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(id)
	browser.Open(url)
}

func OpenLinkInBrowser(list *cview.List, appState *structs.ApplicationState, submissions []*structs.Submission) {
	item := list.GetCurrentItemIndex() + appState.SubmissionsToShow*(appState.CurrentPage)
	url := submissions[item].URL
	browser.Open(url)
}

func NextPage(
	app *cview.Application,
	list *cview.List,
	submissions *structs.Submissions,
	main *structs.MainView,
	appState *structs.ApplicationState) {

	nextPage := appState.CurrentPage + 1

	if nextPage > submissions.MaxPages {
		return
	}

	currentlySelectedItem := list.GetCurrentItemIndex()

	if !pageHasEnoughSubmissionsToView(nextPage, appState.SubmissionsToShow, submissions.Entries) {
		err := fetchAndAppendSubmissionEntries(submissions, appState)
		if err != nil {
			setApplicationToErrorState(appState, main, list, app)
			return
		}
	}

	appState.CurrentPage++

	SetListItemsToCurrentPage(list, submissions.Entries, appState.CurrentPage, appState.SubmissionsToShow)
	list.SetCurrentItem(currentlySelectedItem)

	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.SubmissionsToShow)
	view.SetPageCounter(main, appState.CurrentPage, submissions.MaxPages, "orange")
}

func pageHasEnoughSubmissionsToView(page int, visibleStories int, submissions []*structs.Submission) bool {
	largestItemToDisplay := (page * visibleStories) + visibleStories
	downloadedSubmissions := len(submissions)

	return downloadedSubmissions > largestItemToDisplay
}

func fetchAndAppendSubmissionEntries(submissions *structs.Submissions, appState *structs.ApplicationState) error {
	submissions.PageToFetchFromAPI++
	submissionEntries, err := fetcher.FetchSubmissionEntries(submissions.PageToFetchFromAPI, appState.SubmissionsCategory)
	submissions.Entries = append(submissions.Entries, submissionEntries...)
	return err
}

func SetListItemsToCurrentPage(list *cview.List, submissions []*structs.Submission, currentPage int, viewableStories int) {
	list.Clear()
	start := currentPage * viewableStories
	end := start + viewableStories

	for i := start; i < end; i++ {
		s := submissions[i]
		mainText := formatter.GetMainText(s.Title, s.Domain)
		secondaryText := formatter.GetSecondaryText(s.Points, s.Author, s.Time, s.CommentsCount)

		item := cview.NewListItem(mainText)
		item.SetSecondaryText(secondaryText)

		list.AddItem(item)
	}
}

func ChangeCategory(
	app *cview.Application,
	event *tcell.EventKey,
	list *cview.List,
	appState *structs.ApplicationState,
	submissions []*structs.Submissions,
	main *structs.MainView) {
	currentItem := list.GetCurrentItemIndex()
	if event.Key() == tcell.KeyBacktab {
		appState.SubmissionsCategory = getPreviousCategory(appState.SubmissionsCategory, 4)
	} else {
		appState.SubmissionsCategory = getNextCategory(appState.SubmissionsCategory, 4)
	}

	currentSubmissions := submissions[appState.SubmissionsCategory]
	appState.CurrentPage = 0

	if !pageHasEnoughSubmissionsToView(0, appState.SubmissionsToShow, currentSubmissions.Entries) {
		err := fetchAndAppendSubmissionEntries(currentSubmissions, appState)
		if err != nil {
			setApplicationToErrorState(appState, main, list, app)
			return
		}
	}

	SetListItemsToCurrentPage(list, currentSubmissions.Entries, appState.CurrentPage, appState.SubmissionsToShow)
	list.SetCurrentItem(currentItem)

	view.SetPageCounter(main, appState.CurrentPage, currentSubmissions.MaxPages, "orange")
	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.SubmissionsToShow)
	view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.SubmissionsCategory)
}

func getNextCategory(currentCategory int, numberOfCategories int) int {
	if currentCategory == (numberOfCategories - 1) {
		return 0
	} else {
		return currentCategory + 1
	}
}

func getPreviousCategory(currentCategory int, numberOfCategories int) int {
	if currentCategory == 0 {
		return numberOfCategories - 1
	} else {
		return currentCategory - 1
	}
}

func ChangeHelpScreenCategory(event *tcell.EventKey, appState *structs.ApplicationState, main *structs.MainView) {
	if event.Key() == tcell.KeyBacktab {
		appState.HelpScreenCategory = getPreviousCategory(appState.HelpScreenCategory, 3)
	} else {
		appState.HelpScreenCategory = getNextCategory(appState.HelpScreenCategory, 3)
	}

	showInfoCategory(main, appState)
}

func PreviousPage(list *cview.List, submissions *structs.Submissions, main *structs.MainView, appState *structs.ApplicationState) {
	previousPage := appState.CurrentPage - 1
	if previousPage < 0 {
		return
	}

	appState.CurrentPage--
	currentlySelectedItem := list.GetCurrentItemIndex()

	SetListItemsToCurrentPage(list, submissions.Entries, appState.CurrentPage, appState.SubmissionsToShow)

	list.SetCurrentItem(currentlySelectedItem)

	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.SubmissionsToShow)
	view.SetPageCounter(main, appState.CurrentPage, submissions.MaxPages, "orange")
}

func SelectNextSettingsElement(list *cview.List) {
	currentItem := list.GetCurrentItemIndex()
	itemCount := list.GetItemCount()
	unselectableElements := settings.GetUnselectableItems()

	if currentItem == itemCount {
		return
	}

	next := currentItem + 1
	for intInSlice(next, unselectableElements) {
		if next == itemCount {
			return
		}
		next++
	}

	list.SetCurrentItem(next)
}

func ShowModal(main *structs.MainView) {
	main.Panels.ShowPanel(constants.ModalPanel)
}

func SelectPreviousSettingsElement(list *cview.List) {
	currentItem := list.GetCurrentItemIndex()
	unselectableElements := settings.GetUnselectableItems()

	if currentItem == 0 {
		return
	}

	prev := currentItem - 1
	for intInSlice(prev, unselectableElements) {
		if prev == 0 {
			return
		}
		prev--
	}

	list.SetCurrentItem(prev)
}

func SelectNextSettingsCategory(app *cview.Application,
	list *cview.List,
	settings *structs.Settings,
	submissions []*structs.Submissions,
	main *structs.MainView,
	appState *structs.ApplicationState) {
	//getNextCategory()
}

func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func SelectNextElement(list *cview.List) {
	currentItem := list.GetCurrentItemIndex()
	itemCount := list.GetItemCount()

	if currentItem == itemCount {
		return
	} else {
		list.SetCurrentItem(currentItem + 1)
	}
}

func SelectPreviousElement(list *cview.List) {
	currentItem := list.GetCurrentItemIndex()

	if currentItem == 0 {
		return
	} else {
		list.SetCurrentItem(currentItem - 1)
	}
}

func EnterInfoScreen(main *structs.MainView, appState *structs.ApplicationState) {
	appState.IsOnHelpScreen = true

	showInfoCategory(main, appState)
}

func showInfoCategory(main *structs.MainView, appState *structs.ApplicationState) {
	if appState.HelpScreenCategory == constants.Settings {
		view.SetPageCounter(main, 0, 1, "#82aaff")
	} else {
		view.HidePageCounter(main)
	}

	view.SetHelpScreenHeader(main, appState.ScreenWidth, appState.HelpScreenCategory)
	view.HideLeftMarginRanks(main)
	view.SetHelpScreenPanel(main, appState.HelpScreenCategory)
}

func ExitHelpScreen(main *structs.MainView, appState *structs.ApplicationState, submissions *structs.Submissions) {
	appState.IsOnHelpScreen = false

	view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.SubmissionsCategory)
	view.SetPanelToSubmissions(main)
	view.SetPageCounter(main, appState.CurrentPage, submissions.MaxPages, "orange")
	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.SubmissionsToShow)
	view.HideStatusBar(main)
}

func SelectFirstElementInList(list *cview.List) {
	view.SelectFirstElementInList(list)
}

func SelectLastElementInList(list *cview.List) {
	view.SelectLastElementInList(list)
}

func SelectElementInList(list *cview.List, element rune) {
	i := element - '0'
	adjustedIndex := int(i) - 1

	if int(i) == 0 {
		tenthElement := 9
		view.SelectElementInList(list, tenthElement)
	} else {
		view.SelectElementInList(list, adjustedIndex)
	}
}

func Quit(app *cview.Application) {
	app.Stop()
}

func Refresh(app *cview.Application,
	list *cview.List,
	main *structs.MainView,
	submissions []*structs.Submissions,
	appState *structs.ApplicationState) {
	afterResizeFunc := app.GetAfterResizeFunc()
	afterResizeFunc(appState.ScreenWidth, appState.ScreenHeight)

	ExitHelpScreen(main, appState, submissions[appState.SubmissionsCategory])

	if appState.IsOffline {
		list.Clear()
		view.SetPermanentStatusBar(main, constants.OfflineMessage)
		app.Draw()
	} else {
		duration := time.Millisecond * 2000
		view.SetTemporaryStatusBar(app, main, "Refreshed", duration)
	}
}
