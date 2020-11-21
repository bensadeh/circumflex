package model

import (
	"clx/browser"
	"clx/cli"
	cp "clx/comment-parser"
	"clx/http"
	"clx/submission/fetcher"
	formatter2 "clx/submission/formatter"
	"clx/types"
	"clx/view"
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"strconv"
)

func NextPage(app *cview.Application,
	subState *types.SubmissionState,
	main *types.MainView,
	appState *types.ApplicationState) {

	nextPage := subState.CurrentPage + 1

	if nextPage > subState.MaxPages {
		return
	}

	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(main.Panels)

	list := GetListFromFrontPanel(main.Panels)

	if !pageHasEnoughSubmissionsToView(nextPage, subState.ViewableStoriesOnSinglePage, subState.Submissions) {
		fetchAndAppendSubmissions(subState, appState)
	}

	SetList(list, subState.Submissions, nextPage, subState.ViewableStoriesOnSinglePage, app)
	list.SetCurrentItem(currentlySelectedItem)

	subState.CurrentPage++

	view.SetLeftMarginRanks(main, subState.CurrentPage, subState.ViewableStoriesOnSinglePage)
	view.SetFooterText(main, subState.CurrentPage, appState.ScreenWidth, subState.MaxPages)
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

func fetchAndAppendSubmissions(state *types.SubmissionState, cat *types.ApplicationState) {
	newSubs, _ := FetchSubmissions(state, cat)
	state.Submissions = append(state.Submissions, newSubs...)
}

func FetchSubmissions(state *types.SubmissionState, cat *types.ApplicationState) ([]*types.Submission, error) {
	state.PageToFetchFromAPI++
	return fetcher.FetchSubmissions(state.PageToFetchFromAPI, cat.CurrentCategory)
}

func SetList(list *cview.List, submissions []*types.Submission, page int, submissionsToShow int, app *cview.Application) {
	list.Clear()
	start := page * submissionsToShow
	end := start + submissionsToShow

	for i := start; i < end; i++ {
		s := submissions[i]
		mainText := formatter2.GetMainText(s.Title, s.Domain)
		secondaryText := formatter2.GetSecondaryText(s.Points, s.Author, s.Time, s.CommentsCount)

		item := cview.NewListItem(mainText)
		item.SetSecondaryText(secondaryText)

		list.AddItem(item)
	}

	SetSelectedFunction(app, list, submissions, page, submissionsToShow)
}

func SetSelectedFunction(
	app *cview.Application,
	list *cview.List,
	submissions []*types.Submission,
	currentPage int,
	viewableStories int) {

	list.SetSelectedFunc(func(i int, _ *cview.ListItem) {
		app.Suspend(func() {
			for index := range submissions {
				if index == i {
					storyIndex := (currentPage)*viewableStories + i
					s := submissions[storyIndex]

					if s.Author == "" {
						return
					}

					id := strconv.Itoa(s.ID)
					JSON, _ := http.Get("http://node-hnapi.herokuapp.com/item/" + id)
					jComments := new(cp.Comments)
					_ = json.Unmarshal(JSON, jComments)

					commentTree := cp.PrintCommentTree(*jComments, 4, 70)
					cli.Less(commentTree)
				}
			}
		})
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'o' {
			item := list.GetCurrentItemIndex() + viewableStories*(currentPage)
			url := submissions[item].URL
			browser.Open(url)
		} else if event.Rune() == 'c' {
			item := list.GetCurrentItemIndex() + viewableStories*(currentPage)
			id := submissions[item].ID
			url := "https://news.ycombinator.com/item?id=" + strconv.Itoa(id)
			browser.Open(url)
		}
		return event
	})
}

func ChangeCategory(event *tcell.EventKey,
	appState *types.ApplicationState,
	subStates []*types.SubmissionState,
	main *types.MainView,
	app *cview.Application) {
	if event.Key() == tcell.KeyBacktab {
		appState.CurrentCategory = getPreviousCategory(appState.CurrentCategory)
	} else {
		appState.CurrentCategory = getNextCategory(appState.CurrentCategory)
	}

	nextState := subStates[appState.CurrentCategory]
	nextState.CurrentPage = 0

	if !pageHasEnoughSubmissionsToView(0, nextState.ViewableStoriesOnSinglePage, nextState.Submissions) {
		fetchAndAppendSubmissions(nextState, appState)
	}

	view.SetPanelCategory(main, appState.CurrentCategory)
	list := GetListFromFrontPanel(main.Panels)
	SetList(list, nextState.Submissions, 0, nextState.ViewableStoriesOnSinglePage, app)

	view.SetFooterText(main, nextState.CurrentPage, appState.ScreenWidth, nextState.MaxPages)
	view.SetLeftMarginRanks(main, nextState.CurrentPage, nextState.ViewableStoriesOnSinglePage)
	view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.CurrentCategory)
}

func getNextCategory(currentCategory int) int {
	switch currentCategory {
	case types.NoCategory:
		return types.New
	case types.New:
		return types.Ask
	case types.Ask:
		return types.Show
	case types.Show:
		return types.NoCategory
	default:
		return 0
	}
}

func getPreviousCategory(currentCategory int) int {
	switch currentCategory {
	case types.NoCategory:
		return types.Show
	case types.Show:
		return types.Ask
	case types.Ask:
		return types.New
	case types.New:
		return types.NoCategory
	default:
		return 0
	}
}

func PreviousPage(app *cview.Application,
	state *types.SubmissionState,
	main *types.MainView,
	appState *types.ApplicationState) {

	previousPage := state.CurrentPage - 1
	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(main.Panels)

	if previousPage < 0 {
		return
	}

	list := GetListFromFrontPanel(main.Panels)

	SetList(list, state.Submissions, previousPage, state.ViewableStoriesOnSinglePage, app)
	list.SetCurrentItem(currentlySelectedItem)

	state.CurrentPage--

	view.SetLeftMarginRanks(main, state.CurrentPage, state.ViewableStoriesOnSinglePage)
	view.SetFooterText(main, state.CurrentPage, appState.ScreenWidth, state.MaxPages)
}

func ShowHelpScreen(main *types.MainView, screenWidth int) {
	view.SetKeymapsHeader(main, screenWidth)
	view.HideLeftMarginRanks(main)
	view.HideFooterText(main)
	view.SetPanelToHelpScreen(main)
}

func ReturnFromHelpScreen(main *types.MainView, screenWidth int, cat *types.ApplicationState, currentPage int, currentState *types.SubmissionState, viewableStories int) {
	view.SetHackerNewsHeader(main, screenWidth, cat.CurrentCategory)
	view.SetPanelCategory(main, cat.CurrentCategory)
	view.SetFooterText(main, currentPage, screenWidth, currentState.MaxPages)
	view.SetLeftMarginRanks(main, currentPage, viewableStories)
}

func SelectLastElementInList(currentState *types.SubmissionState, main *types.MainView) {
	view.SelectLastElementInList(currentState, main)
}

func SelectFirstElementInList(main *types.MainView) {
	view.SelectFirstElementInList(main)
}
