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

	setList(list, state.Submissions, nextPage, state.ViewableStoriesOnSinglePage, app)
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
	newSubs, _ := fetchSubmissions(state, cat)
	state.Submissions = append(state.Submissions, newSubs...)
}

func fetchSubmissions(state *types.ApplicationState, cat *types.Category) ([]*types.Submission, error) {
	state.PageToFetchFromAPI++
	return fetcher.FetchSubmissions(state.PageToFetchFromAPI, cat.CurrentCategory)
}

func setList(list *cview.List, submissions []*types.Submission, page int, submissionsToShow int, app *cview.Application) {
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
