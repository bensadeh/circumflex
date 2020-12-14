package model

import (
	"clx/browser"
	"clx/cli"
	cp "clx/comment-parser"
	"clx/http"
	"clx/screen"
	"clx/submission/fetcher"
	formatter2 "clx/submission/formatter"
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
					storyIndex := (appState.CurrentPage)*appState.ViewableStoriesOnSinglePage + i
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
			item := list.GetCurrentItemIndex() + appState.ViewableStoriesOnSinglePage*(appState.CurrentPage)
			url := submissions[item].URL
			browser.Open(url)
		} else if event.Rune() == 'c' {
			item := list.GetCurrentItemIndex() + appState.ViewableStoriesOnSinglePage*(appState.CurrentPage)
			id := submissions[item].ID
			url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(id)
			browser.Open(url)
		}
		return event
	})
}

func NextPage(
	app *cview.Application,
	submissions *types.Submissions,
	main *types.MainView,
	appState *types.ApplicationState) {
	nextPage := appState.CurrentPage + 1

	if nextPage > submissions.MaxPages {
		return
	}

	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(main.Panels)

	list := GetListFromFrontPanel(main.Panels)

	if !pageHasEnoughSubmissionsToView(nextPage, appState.ViewableStoriesOnSinglePage, submissions.SubmissionEntries) {
		FetchAndAppendSubmissions(submissions, appState)
	}

	appState.CurrentPage++

	SetListItemsToCurrentPage(list, submissions.SubmissionEntries, appState)
	SetShortcutsForListItems(app, list, submissions.SubmissionEntries, appState)
	list.SetCurrentItem(currentlySelectedItem)

	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.ViewableStoriesOnSinglePage)
	view.SetFooter(main, appState.CurrentPage, appState.ScreenWidth, submissions.MaxPages)
}

func getCurrentlySelectedItemOnFrontPage(pages *cview.Panels) int {
	_, primitive := pages.GetFrontPanel()
	list, ok := primitive.(*cview.List)
	if ok {
		return list.GetCurrentItemIndex()
	}
	return 0
}

func GetListFromFrontPanel(pages *cview.Panels) *cview.List {
	_, primitive := pages.GetFrontPanel()
	list, _ := primitive.(*cview.List)
	return list
}

func pageHasEnoughSubmissionsToView(page int, visibleStories int, submissions []*types.Submission) bool {
	largestItemToDisplay := (page * visibleStories) + visibleStories
	downloadedSubmissions := len(submissions)

	return downloadedSubmissions > largestItemToDisplay
}

func FetchAndAppendSubmissions(state *types.Submissions, appState *types.ApplicationState) {
	newSubs, _ := FetchSubmissions(state, appState)
	state.SubmissionEntries = append(state.SubmissionEntries, newSubs...)
}

func FetchSubmissions(state *types.Submissions, appState *types.ApplicationState) ([]*types.Submission, error) {
	state.PageToFetchFromAPI++
	return fetcher.FetchSubmissions(state.PageToFetchFromAPI, appState.CurrentCategory)
}

func SetListItemsToCurrentPage(list *cview.List, submissions []*types.Submission, appState *types.ApplicationState) {
	list.Clear()
	start := appState.CurrentPage * appState.ViewableStoriesOnSinglePage
	end := start + appState.ViewableStoriesOnSinglePage

	for i := start; i < end; i++ {
		s := submissions[i]
		mainText := formatter2.GetMainText(s.Title, s.Domain)
		secondaryText := formatter2.GetSecondaryText(s.Points, s.Author, s.Time, s.CommentsCount)

		item := cview.NewListItem(mainText)
		item.SetSecondaryText(secondaryText)

		list.AddItem(item)
	}
}

func ChangeCategory(
	event *tcell.EventKey,
	appState *types.ApplicationState,
	submissions []*types.Submissions,
	main *types.MainView,
	app *cview.Application) {
	if event.Key() == tcell.KeyBacktab {
		appState.CurrentCategory = getPreviousCategory(appState.CurrentCategory)
	} else {
		appState.CurrentCategory = getNextCategory(appState.CurrentCategory)
	}

	nextState := submissions[appState.CurrentCategory]
	appState.CurrentPage = 0

	if !pageHasEnoughSubmissionsToView(0, appState.ViewableStoriesOnSinglePage, nextState.SubmissionEntries) {
		FetchAndAppendSubmissions(nextState, appState)
	}

	view.SetPanelCategory(main, appState.CurrentCategory)
	list := GetListFromFrontPanel(main.Panels)
	SetListItemsToCurrentPage(list, nextState.SubmissionEntries, appState)
	SetShortcutsForListItems(app, list, nextState.SubmissionEntries, appState)

	view.SetFooter(main, appState.CurrentPage, appState.ScreenWidth, nextState.MaxPages)
	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.ViewableStoriesOnSinglePage)
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
	submissions *types.Submissions,
	main *types.MainView,
	appState *types.ApplicationState) {
	previousPage := appState.CurrentPage - 1
	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(main.Panels)

	if previousPage < 0 {
		return
	}

	list := GetListFromFrontPanel(main.Panels)

	appState.CurrentPage--

	SetListItemsToCurrentPage(list, submissions.SubmissionEntries, appState)
	SetShortcutsForListItems(app, list, submissions.SubmissionEntries, appState)
	list.SetCurrentItem(currentlySelectedItem)

	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.ViewableStoriesOnSinglePage)
	view.SetFooter(main, appState.CurrentPage, appState.ScreenWidth, submissions.MaxPages)
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
	view.SetPanelCategory(main, appState.CurrentCategory)
	view.SetFooter(main, appState.CurrentPage, appState.ScreenWidth, submissions.MaxPages)
	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.ViewableStoriesOnSinglePage)
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
	view.SetPanelCategory(main, appState.CurrentCategory)
	view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.CurrentCategory)
	view.SetLeftMarginRanks(main, 0, appState.ViewableStoriesOnSinglePage)
	view.SetFooter(main,
		0,
		appState.ScreenWidth,
		submissions[appState.CurrentCategory].MaxPages)
}

func ShowPageAfterResize(
	appState *types.ApplicationState,
	submissions []*types.Submissions,
	main *types.MainView,
	app *cview.Application) {
	frontPanelList := GetListFromFrontPanel(main.Panels)
	submissionEntries := submissions[appState.CurrentCategory].SubmissionEntries

	SetListItemsToCurrentPage(frontPanelList, submissionEntries, appState)
	SetShortcutsForListItems(app, frontPanelList, submissionEntries, appState)

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
	appState.ViewableStoriesOnSinglePage = screen.GetViewableStoriesOnSinglePage(
		appState.ScreenHeight,
		30)
}

func resetSubmissionStates(submissions []*types.Submissions) {
	numberOfCategories := 3

	for i := 0; i < numberOfCategories; i++ {
		submissions[i].MappedSubmissions = 0
		submissions[i].PageToFetchFromAPI = 0
		submissions[i].StoriesListed = 0
		submissions[i].SubmissionEntries = nil
	}
}
