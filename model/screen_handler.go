package model

import (
	"clx/browser"
	"clx/cli"
	cp "clx/comment-parser"
	"clx/http"
	"clx/screen"
	"clx/submission/fetcher"
	"clx/submission/formatter"
	"clx/types"
	"clx/view"
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"strconv"
	"time"
)

func ReadSubmissionComments(
	app *cview.Application,
	list *cview.List,
	submissions []*types.Submission,
	appState *types.ApplicationState) {
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

func OpenCommentsInBrowser(list *cview.List, appState *types.ApplicationState, submissions []*types.Submission) {
	item := list.GetCurrentItemIndex() + appState.SubmissionsToShow*(appState.CurrentPage)
	id := submissions[item].ID
	url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(id)
	browser.Open(url)
}

func OpenLinkInBrowser(list *cview.List, appState *types.ApplicationState, submissions []*types.Submission) {
	item := list.GetCurrentItemIndex() + appState.SubmissionsToShow*(appState.CurrentPage)
	url := submissions[item].URL
	browser.Open(url)
}

func NextPage(list *cview.List, submissions *types.Submissions, main *types.MainView, appState *types.ApplicationState) {
	nextPage := appState.CurrentPage + 1

	if nextPage > submissions.MaxPages {
		return
	}

	currentlySelectedItem := list.GetCurrentItemIndex()

	if !pageHasEnoughSubmissionsToView(nextPage, appState.SubmissionsToShow, submissions.Entries) {
		FetchAndAppendSubmissionEntries(submissions, appState)
	}

	appState.CurrentPage++

	SetListItemsToCurrentPage(list, submissions.Entries, appState.CurrentPage, appState.SubmissionsToShow)
	list.SetCurrentItem(currentlySelectedItem)

	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.SubmissionsToShow)
	view.SetPageCounter(main, appState.CurrentPage, submissions.MaxPages)
}

func pageHasEnoughSubmissionsToView(page int, visibleStories int, submissions []*types.Submission) bool {
	largestItemToDisplay := (page * visibleStories) + visibleStories
	downloadedSubmissions := len(submissions)

	return downloadedSubmissions > largestItemToDisplay
}

func FetchAndAppendSubmissionEntries(submissions *types.Submissions, appState *types.ApplicationState) {
	submissions.PageToFetchFromAPI++
	submissionEntries, _ := fetcher.FetchSubmissionEntries(submissions.PageToFetchFromAPI, appState.CurrentCategory)
	submissions.Entries = append(submissions.Entries, submissionEntries...)
}

func SetListItemsToCurrentPage(list *cview.List, submissions []*types.Submission, currentPage int, viewableStories int) {
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
	event *tcell.EventKey,
	list *cview.List,
	appState *types.ApplicationState,
	submissions []*types.Submissions,
	main *types.MainView) {
	currentItem := list.GetCurrentItemIndex()
	if event.Key() == tcell.KeyBacktab {
		appState.CurrentCategory = getPreviousCategory(appState.CurrentCategory)
	} else {
		appState.CurrentCategory = getNextCategory(appState.CurrentCategory)
	}

	currentSubmissions := submissions[appState.CurrentCategory]
	appState.CurrentPage = 0

	if !pageHasEnoughSubmissionsToView(0, appState.SubmissionsToShow, currentSubmissions.Entries) {
		FetchAndAppendSubmissionEntries(currentSubmissions, appState)
	}

	SetListItemsToCurrentPage(list, currentSubmissions.Entries, appState.CurrentPage, appState.SubmissionsToShow)
	list.SetCurrentItem(currentItem)

	view.SetPageCounter(main, appState.CurrentPage, currentSubmissions.MaxPages)
	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.SubmissionsToShow)
	view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.CurrentCategory)
}

func getNextCategory(currentCategory int) int {
	lastCategory := types.Show
	firstCategory := types.FrontPage

	if currentCategory == lastCategory {
		return firstCategory
	} else {
		return currentCategory + 1
	}
}

func getPreviousCategory(currentCategory int) int {
	lastCategory := types.Show
	firstCategory := types.FrontPage

	if currentCategory == firstCategory {
		return lastCategory
	} else {
		return currentCategory - 1
	}
}

func PreviousPage(list *cview.List, submissions *types.Submissions, main *types.MainView, appState *types.ApplicationState) {
	previousPage := appState.CurrentPage - 1
	if previousPage < 0 {
		return
	}

	appState.CurrentPage--
	currentlySelectedItem := list.GetCurrentItemIndex()

	SetListItemsToCurrentPage(list, submissions.Entries, appState.CurrentPage, appState.SubmissionsToShow)

	list.SetCurrentItem(currentlySelectedItem)

	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.SubmissionsToShow)
	view.SetPageCounter(main, appState.CurrentPage, submissions.MaxPages)
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

func ShowHelpScreen(main *types.MainView, appState *types.ApplicationState) {
	appState.IsOnHelpScreen = true

	view.SetKeymapsHeader(main, appState.ScreenWidth)
	view.HideLeftMarginRanks(main)
	view.HideFooterText(main)
	view.SetPanelToHelpScreen(main)
}

func ReturnFromHelpScreen(main *types.MainView, appState *types.ApplicationState, submissions *types.Submissions) {
	appState.IsOnHelpScreen = false

	view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.CurrentCategory)
	view.SetPanelToSubmissions(main)
	view.SetPageCounter(main, appState.CurrentPage, submissions.MaxPages)
	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.SubmissionsToShow)
}

func SelectFirstElementInList(main *types.MainView) {
	view.SelectFirstElementInList(main)
}

func SelectLastElementInList(main *types.MainView, appState *types.ApplicationState) {
	view.SelectLastElementInList(main, appState)
}

func SelectElementInList(main *types.MainView, element rune) {
	i := element - '0'
	adjustedIndex := int(i) - 1

	if int(i) == 0 {
		tenthElement := 9
		view.SelectElementInList(main, tenthElement)
	} else {
		view.SelectElementInList(main, adjustedIndex)
	}
}

func InitializeHeaderAndFooterAndLeftMarginView(
	appState *types.ApplicationState,
	submissions []*types.Submissions,
	main *types.MainView) {
	view.SetPanelToSubmissions(main)
	view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.CurrentCategory)
	view.SetLeftMarginRanks(main, 0, appState.SubmissionsToShow)
	view.SetPageCounter(main, 0, submissions[appState.CurrentCategory].MaxPages)
}

func ShowPageAfterResize(
	appState *types.ApplicationState,
	list *cview.List,
	submissions []*types.Submissions,
	main *types.MainView) {
	submissionEntries := submissions[appState.CurrentCategory].Entries

	SetListItemsToCurrentPage(list, submissionEntries, appState.CurrentPage, appState.SubmissionsToShow)

	if appState.IsOnHelpScreen {
		ShowHelpScreen(main, appState)
	}
}

func Quit(app *cview.Application) {
	app.Stop()
}

func Refresh(app *cview.Application, main *types.MainView, appState *types.ApplicationState) {
	afterResizeFunc := app.GetAfterResizeFunc()
	afterResizeFunc(appState.ScreenWidth, appState.ScreenHeight)
	duration := time.Millisecond * 2000
	view.SetTemporaryStatusBar(app, main, "Refreshed", duration)
}

func ResetStates(appState *types.ApplicationState, submissions []*types.Submissions) {
	resetApplicationState(appState)
	resetSubmissionStates(submissions)
}

func resetApplicationState(appState *types.ApplicationState) {
	appState.CurrentPage = 0
	appState.ScreenWidth = screen.GetTerminalWidth()
	appState.ScreenHeight = screen.GetTerminalHeight()
	appState.SubmissionsToShow = screen.GetSubmissionsToShow(appState.ScreenHeight, 30)
}

func resetSubmissionStates(submissions []*types.Submissions) {
	numberOfCategories := 3

	for i := 0; i < numberOfCategories; i++ {
		submissions[i].MappedSubmissions = 0
		submissions[i].PageToFetchFromAPI = 0
		submissions[i].StoriesListed = 0
		submissions[i].Entries = nil
	}
}
