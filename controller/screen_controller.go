package controller

import (
	"clx/browser"
	"clx/cli"
	cp "clx/comment-parser"
	"clx/http"
	"clx/primitives"
	"clx/screen"
	"clx/submission/fetcher"
	formatter2 "clx/submission/formatter"
	"clx/types"
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"strconv"
)

const (
	maximumStoriesToDisplay = 30
	helpPage                = "help"
	offlinePage             = "offline"
	maxPages                = 3
)

type screenController struct {
	Submissions      *types.Subs
	Application      *cview.Application
	MainView         *primitives.MainView
	ApplicationState *types.ApplicationState
}

func NewScreenController() *screenController {
	sc := new(screenController)
	sc.Application = cview.NewApplication()
	sc.ApplicationState = new(types.ApplicationState)
	sc.Submissions = new(types.Subs)
	sc.ApplicationState.MaxPages = maxPages
	sc.ApplicationState.ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState.ScreenHeight = screen.GetTerminalHeight()
	sc.ApplicationState.ViewableStoriesOnSinglePage = screen.GetViewableStoriesOnSinglePage(
		sc.ApplicationState.ScreenHeight,
		maximumStoriesToDisplay)

	sc.MainView = primitives.NewMainView(
		sc.ApplicationState.ScreenWidth,
		sc.ApplicationState.ViewableStoriesOnSinglePage)

	newSubmissions, err := fetchSubmissions(sc.ApplicationState)
	sc.ApplicationState.IsOffline = getIsOfflineStatus(err)

	mapSubmissions(sc.Application,
		sc.ApplicationState,
		sc.Submissions,
		newSubmissions,
		sc.MainView)

	startPage := getStartPage(sc.ApplicationState.IsOffline)
	sc.MainView.Pages.SwitchToPage(startPage)

	setShortcuts(sc.Application, sc.ApplicationState, sc.MainView, sc.Submissions)

	return sc
}

func getIsOfflineStatus(err error) bool {
	if err != nil {
		return true
	}
	return false
}

func getStartPage(isOffline bool) string {
	if isOffline {
		return "offline"
	}
	return "0"
}

func (sc *screenController) getCurrentPage() string {
	return strconv.Itoa(sc.ApplicationState.CurrentPage)
}

func setShortcuts(app *cview.Application,
	state *types.ApplicationState,
	main *primitives.MainView,
	oldSubmissions *types.Subs) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentPage, _ := main.Pages.GetFrontPage()

		if currentPage == offlinePage {
			if event.Rune() == 'q' {
				app.Stop()
			}
			return event
		}

		if currentPage == helpPage {
			main.SetHeaderTextToHN(state.ScreenWidth)
			p := strconv.Itoa(state.CurrentPage)
			main.Pages.SwitchToPage(p)
			main.SetFooterText(state.CurrentPage, state.ScreenWidth)
			main.SetLeftMarginRanks(state.CurrentPage, state.ViewableStoriesOnSinglePage)
			return event
		}

		if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			nextPage(main.Pages,
				state,
				app,
				oldSubmissions,
				main)
			main.SetLeftMarginRanks(state.CurrentPage,
				state.ViewableStoriesOnSinglePage)
			main.SetFooterText(state.CurrentPage,
				state.ScreenWidth)
		} else if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			previousPage(state, main.Pages)
			main.SetLeftMarginRanks(state.CurrentPage,
				state.ViewableStoriesOnSinglePage)
			main.SetFooterText(state.CurrentPage,
				state.ScreenWidth)
		} else if event.Rune() == 'q' {
			app.Stop()
		} else if event.Rune() == 'i' || event.Rune() == '?' {
			main.SetHeaderTextToKeymaps(state.ScreenWidth)
			main.HideFooterText()
			main.HideLeftMarginRanks()
			main.Pages.SwitchToPage(helpPage)
		}
		return event
	})
}

func nextPage(pages *cview.Pages,
	state *types.ApplicationState,
	app *cview.Application,
	oldSubmissions *types.Subs,
	main *primitives.MainView) {
	nextPage := state.CurrentPage + 1

	if nextPage > maxPages {
		return
	}

	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(pages)

	if nextPage < state.MappedPages {
		pages.SwitchToPage(strconv.Itoa(nextPage))
		app.ForceDraw()
		setCurrentlySelectedItemOnFrontPage(currentlySelectedItem, pages)
	} else {
		newSubmissions, _ := fetchSubmissions(state)
		mapSubmissions(app,
			state,
			oldSubmissions,
			newSubmissions,
			main)
		pages.SwitchToPage(strconv.Itoa(nextPage))

		app.ForceDraw()
		setCurrentlySelectedItemOnFrontPage(currentlySelectedItem, pages)
	}

	state.CurrentPage++

}

func getCurrentlySelectedItemOnFrontPage(pages *cview.Pages) int {
	_, primitive := pages.GetFrontPage()
	list := primitive.(*cview.List)
	return list.GetCurrentItemIndex()
}

func setCurrentlySelectedItemOnFrontPage(item int, pages *cview.Pages) {
	_, primitive := pages.GetFrontPage()
	list := primitive.(*cview.List)
	list.SetCurrentItem(item)
}

func previousPage(state *types.ApplicationState, pages *cview.Pages) {
	previousPage := state.CurrentPage - 1

	if previousPage < 0 {
		return
	}

	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(pages)

	state.CurrentPage--
	pages.SwitchToPage(strconv.Itoa(state.CurrentPage))

	setCurrentlySelectedItemOnFrontPage(currentlySelectedItem, pages)
}

func (sc *screenController) getStoriesToDisplay() int {
	return sc.ApplicationState.ViewableStoriesOnSinglePage
}

func setSelectedFunction(app *cview.Application,
	list *cview.List,
	submissions *types.Subs,
	state *types.ApplicationState) {
	list.SetSelectedFunc(func(i int, _ *cview.ListItem) {
		app.Suspend(func() {
			for index := range submissions.News {
				if index == i {
					storyIndex := (state.CurrentPage)*state.ViewableStoriesOnSinglePage + i
					s := submissions.News[storyIndex]
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
			item := list.GetCurrentItemIndex() + state.ViewableStoriesOnSinglePage*(state.CurrentPage)
			url := submissions.News[item].URL
			browser.Open(url)
		}
		return event
	})
}

func fetchSubmissions(state *types.ApplicationState) ([]*types.Submission, error) {
	state.PageToFetchFromAPI++
	return fetcher.FetchSubmissions(state.PageToFetchFromAPI)
}

func mapSubmissions(app *cview.Application,
	state *types.ApplicationState,
	oldSubmissions *types.Subs,
	newSubmissions []*types.Submission,
	main *primitives.MainView) {
	oldSubmissions.News = append(oldSubmissions.News, newSubmissions...)
	mapSubmissionsToListItems(app, oldSubmissions, state, main)
}

func mapSubmissionsToListItems(app *cview.Application,
	oldSubmissions *types.Subs,
	state *types.ApplicationState,
	main *primitives.MainView) {
	for hasStoriesToMap(oldSubmissions.News, state) {
		sub := oldSubmissions.News[state.MappedSubmissions : state.MappedSubmissions+state.ViewableStoriesOnSinglePage]
		list := createNewList(app, oldSubmissions, state)
		addSubmissionsToList(list, sub, state)

		main.Pages.AddPage(strconv.Itoa(state.MappedPages), list, true, true)
		state.MappedPages++
	}
}

func hasStoriesToMap(submissions []*types.Submission, state *types.ApplicationState) bool {
	return len(submissions)-state.MappedSubmissions >= state.ViewableStoriesOnSinglePage
}

func createNewList(app *cview.Application,
	submissions *types.Subs,
	state *types.ApplicationState) *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)
	setSelectedFunction(app, list, submissions, state)

	return list
}

func addSubmissionsToList(list *cview.List,
	submissions []*types.Submission,
	state *types.ApplicationState) {
	for _, s := range submissions {
		mainText := formatter2.GetMainText(s.Title, s.Domain)
		secondaryText := formatter2.GetSecondaryText(s.Points, s.Author, s.Time, s.CommentsCount)

		item := cview.NewListItem(mainText)
		item.SetSecondaryText(secondaryText)

		list.AddItem(item)
		state.MappedSubmissions++
	}
}
