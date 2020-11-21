package model

import (
	"clx/browser"
	"clx/cli"
	cp "clx/comment-parser"
	"clx/http"
	"clx/primitives"
	"clx/submission/fetcher"
	formatter2 "clx/submission/formatter"
	"clx/types"
	"clx/view"
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"strconv"
)

func NextPage(app *cview.Application, state *types.ApplicationState, main *primitives.MainView, cat *types.Category) {
	nextPage := state.CurrentPage + 1

	if nextPage > state.MaxPages {
		return
	}

	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(main.Panels)

	list := getListFromFrontPanel(main.Panels)

	if !pageHasEnoughSubmissionsToView(nextPage, state.ViewableStoriesOnSinglePage, state.Submissions) {
		fetchAndAppendSubmissions(state, cat)
	}

	SetList(list, state.Submissions, nextPage, state.ViewableStoriesOnSinglePage, app)
	list.SetCurrentItem(currentlySelectedItem)

	state.CurrentPage++

	view.SetLeftMarginRanks(main, state.CurrentPage, state.ViewableStoriesOnSinglePage)
	view.SetFooterText(main, state.CurrentPage, state.ScreenWidth, state.MaxPages)
}

func getCurrentlySelectedItemOnFrontPage(pages *cview.Panels) int {
	_, primitive := pages.GetFrontPanel()
	list, ok := primitive.(*cview.List)
	if ok {
		return list.GetCurrentItemIndex()
	}
	return 0
}

func getListFromFrontPanel(pages *cview.Panels) *cview.List {
	_, primitive := pages.GetFrontPanel()
	list, _ := primitive.(*cview.List)
	return list
}

func pageHasEnoughSubmissionsToView(page int, visibleStories int, submissions []*types.Submission) bool {
	largestItemToDisplay := (page * visibleStories) + visibleStories
	downloadedSubmissions := len(submissions)

	return downloadedSubmissions > largestItemToDisplay
}

func fetchAndAppendSubmissions(state *types.ApplicationState, cat *types.Category) {
	newSubs, _ := FetchSubmissions(state, cat)
	state.Submissions = append(state.Submissions, newSubs...)
}

func FetchSubmissions(state *types.ApplicationState, cat *types.Category) ([]*types.Submission, error) {
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

func ChangeCategory(event *tcell.EventKey, cat *types.Category, state []*types.ApplicationState, main *primitives.MainView, app *cview.Application) {
	if event.Key() == tcell.KeyBacktab {
		cat.CurrentCategory = getPreviousCategory(cat.CurrentCategory)
	} else {
		cat.CurrentCategory = getNextCategory(cat.CurrentCategory)
	}

	nextState := state[cat.CurrentCategory]
	nextState.CurrentPage = 0

	if !pageHasEnoughSubmissionsToView(0, nextState.ViewableStoriesOnSinglePage, nextState.Submissions) {
		fetchAndAppendSubmissions(nextState, cat)
	}

	view.SetPanelCategory(main, cat.CurrentCategory)
	list := getListFromFrontPanel(main.Panels)
	SetList(list, nextState.Submissions, 0, nextState.ViewableStoriesOnSinglePage, app)

	view.SetFooterText(main, nextState.CurrentPage, nextState.ScreenWidth, nextState.MaxPages)
	view.SetLeftMarginRanks(main, nextState.CurrentPage, nextState.ViewableStoriesOnSinglePage)
	view.SetHackerNewsHeader(main, nextState.ScreenWidth, cat.CurrentCategory)
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
	state *types.ApplicationState,
	main *primitives.MainView,
	pages *cview.Panels) {

	previousPage := state.CurrentPage - 1
	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(pages)

	if previousPage < 0 {
		return
	}

	list := getListFromFrontPanel(pages)

	SetList(list, state.Submissions, previousPage, state.ViewableStoriesOnSinglePage, app)
	list.SetCurrentItem(currentlySelectedItem)

	state.CurrentPage--

	view.SetLeftMarginRanks(main, state.CurrentPage, state.ViewableStoriesOnSinglePage)
	view.SetFooterText(main, state.CurrentPage, state.ScreenWidth, state.MaxPages)
}

func ShowHelpScreen(main *primitives.MainView, screenWidth int) {
	view.SetKeymapsHeader(main, screenWidth)
	view.HideLeftMarginRanks(main)
	view.HideFooterText(main)
	view.SetPanelToHelpScreen(main)
}

func ReturnFromHelpScreen(main *primitives.MainView, screenWidth int, cat *types.Category, currentPage int, currentState *types.ApplicationState, viewableStories int) {
	view.SetHackerNewsHeader(main, screenWidth, cat.CurrentCategory)
	view.SetPanelCategory(main, cat.CurrentCategory)
	view.SetFooterText(main, currentPage, screenWidth, currentState.MaxPages)
	view.SetLeftMarginRanks(main, currentPage, viewableStories)
}


func SelectLastElementInList(currentState *types.ApplicationState, main *primitives.MainView) {
	view.SelectLastElementInList(currentState, main)
}

func SelectFirstElementInList(main *primitives.MainView) {
	view.SelectFirstElementInList(main)
}
