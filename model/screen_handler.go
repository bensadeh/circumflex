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
)

func SetShortcutsForListItems(
	app *cview.Application,
	list *cview.List,
	submissions []*types.Submission,
	appState *types.ApplicationState) {
	list.SetSelectedFunc(func(i int, _ *cview.ListItem) {
		app.Suspend(func() {
			for index := range submissions {
				if index == i {
					storyIndex := (appState.CurrentPage)*appState.SubmissionsToShow + i
					s := submissions[storyIndex]

					if s.Author == "" {
						appState.IsReturningFromSuspension = true
						return
					}

					id := strconv.Itoa(s.ID)
					JSON, _ := http.Get("http://api.hackerwebapp.com/item/" + id)
					jComments := new(cp.Comments)
					_ = json.Unmarshal(JSON, jComments)

					commentTree := cp.PrintCommentTree(*jComments, 4, 70)
					cli.Less(commentTree)
					appState.IsReturningFromSuspension = true
				}
			}
		})
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'o' {
			item := list.GetCurrentItemIndex() + appState.SubmissionsToShow*(appState.CurrentPage)
			url := submissions[item].URL
			browser.Open(url)
			return event
		}
		if event.Rune() == 'c' {
			item := list.GetCurrentItemIndex() + appState.SubmissionsToShow*(appState.CurrentPage)
			id := submissions[item].ID
			url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(id)
			browser.Open(url)
			return event
		}
		if event.Key() == tcell.KeyTAB {
			list.SetCurrentItem(list.GetCurrentItemIndex())
			return event
		}
		if event.Key() == tcell.KeyBacktab {
			list.SetCurrentItem(list.GetCurrentItemIndex())
			return event
		}

		return event
	})

}

func NextPage(
	app *cview.Application,
	list *cview.List,
	submissions *types.Submissions,
	main *types.MainView,
	appState *types.ApplicationState) {
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
	SetShortcutsForListItems(app, list, submissions.Entries, appState)
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
	main *types.MainView,
	app *cview.Application) {
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
	SetShortcutsForListItems(app, list, currentSubmissions.Entries, appState)
	list.SetCurrentItem(0)

	view.SetPageCounter(main, appState.CurrentPage, currentSubmissions.MaxPages)
	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.SubmissionsToShow)
	view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.CurrentCategory)
}

func getNextCategory(currentCategory int) int {
	switch currentCategory {
	case types.FrontPage:
		return types.New
	case types.New:
		return types.Ask
	case types.Ask:
		return types.Show
	case types.Show:
		return types.FrontPage
	default:
		return 0
	}
}

func getPreviousCategory(currentCategory int) int {
	switch currentCategory {
	case types.FrontPage:
		return types.Show
	case types.Show:
		return types.Ask
	case types.Ask:
		return types.New
	case types.New:
		return types.FrontPage
	default:
		return 0
	}
}

func PreviousPage(
	app *cview.Application,
	list *cview.List,
	submissions *types.Submissions,
	main *types.MainView,
	appState *types.ApplicationState) {
	previousPage := appState.CurrentPage - 1
	if previousPage < 0 {
		return
	}

	appState.CurrentPage--
	currentlySelectedItem := list.GetCurrentItemIndex()

	SetListItemsToCurrentPage(list, submissions.Entries, appState.CurrentPage, appState.SubmissionsToShow)
	SetShortcutsForListItems(app, list, submissions.Entries, appState)

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
		list.SetCurrentItem(currentItem+1)
	}
}

func SelectPreviousElement(list *cview.List) {
	currentItem := list.GetCurrentItemIndex()

	if currentItem == 0 {
		return
	} else {
		list.SetCurrentItem(currentItem-1)
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
	main *types.MainView,
	app *cview.Application) {
	submissionEntries := submissions[appState.CurrentCategory].Entries

	SetListItemsToCurrentPage(list, submissionEntries, appState.CurrentPage, appState.SubmissionsToShow)
	SetShortcutsForListItems(app, list, submissionEntries, appState)

	if appState.IsOnHelpScreen {
		ShowHelpScreen(main, appState)
	}
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
