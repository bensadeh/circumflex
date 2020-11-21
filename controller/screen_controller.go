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
	"clx/view"
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
	"os"
	"strconv"
)

const (
	maximumStoriesToDisplay = 30
	helpPage                = "help"
	offlinePage             = "offline"
)

type screenController struct {
	Application      *cview.Application
	MainView         *primitives.MainView
	ApplicationState []*types.ApplicationState
	Category         *types.Category
}

func NewScreenController() *screenController {
	sc := new(screenController)
	sc.Application = cview.NewApplication()
	sc.ApplicationState = []*types.ApplicationState{}
	sc.ApplicationState = append(sc.ApplicationState, new(types.ApplicationState))
	sc.ApplicationState = append(sc.ApplicationState, new(types.ApplicationState))
	sc.ApplicationState = append(sc.ApplicationState, new(types.ApplicationState))
	sc.ApplicationState = append(sc.ApplicationState, new(types.ApplicationState))
	sc.Category = new(types.Category)
	storiesToDisplay := screen.GetViewableStoriesOnSinglePage(
		screen.GetTerminalHeight(),
		maximumStoriesToDisplay)

	width := screen.GetTerminalWidth()
	height := screen.GetTerminalHeight()

	sc.ApplicationState[types.NoCategory].MaxPages = 2
	sc.ApplicationState[types.NoCategory].ScreenWidth = width
	sc.ApplicationState[types.NoCategory].ScreenHeight = height
	sc.ApplicationState[types.NoCategory].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.ApplicationState[types.New].MaxPages = 2
	sc.ApplicationState[types.New].ScreenWidth = width
	sc.ApplicationState[types.New].ScreenHeight = height
	sc.ApplicationState[types.New].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.ApplicationState[types.Ask].MaxPages = 1
	sc.ApplicationState[types.Ask].ScreenWidth = width
	sc.ApplicationState[types.Ask].ScreenHeight = height
	sc.ApplicationState[types.Ask].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.ApplicationState[types.Show].MaxPages = 1
	sc.ApplicationState[types.Show].ScreenWidth = width
	sc.ApplicationState[types.Show].ScreenHeight = height
	sc.ApplicationState[types.Show].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.MainView = primitives.NewMainView(width, storiesToDisplay)

	newsList := createNewList()
	sc.MainView.Panels.AddPanel(types.NewsPanel, newsList, true, false)
	sc.MainView.Panels.AddPanel(types.NewestPanel, createNewList(), true, false)
	sc.MainView.Panels.AddPanel(types.ShowPanel, createNewList(), true, false)
	sc.MainView.Panels.AddPanel(types.AskPanel, createNewList(), true, false)

	sc.MainView.Panels.SetCurrentPanel(types.NewsPanel)

	view.SetHackerNewsHeader(sc.MainView, width, types.NoCategory)
	view.SetLeftMarginRanks(sc.MainView, 0, storiesToDisplay)
	view.SetFooterText(sc.MainView, 0, width, 2)

	newSubs, err := fetchSubmissions(sc.ApplicationState[types.NoCategory], sc.Category)

	if err != nil {
		println("Error: Could not retrieve submissions")
		os.Exit(1)
	}

	sc.ApplicationState[types.NoCategory].Submissions = append(sc.ApplicationState[types.NoCategory].Submissions, newSubs...)

	setList(newsList, sc.ApplicationState[types.NoCategory].Submissions, 0, storiesToDisplay, sc.Application)

	setShortcuts(sc.Application,
		sc.ApplicationState,
		sc.MainView,
		sc.Category)

	return sc
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

	setSelectedFunction(app, list, submissions, page, submissionsToShow)
}

func fetchAndAppendSubmissions(state *types.ApplicationState, cat *types.Category) {
	newSubs, _ := fetchSubmissions(state, cat)
	state.Submissions = append(state.Submissions, newSubs...)
}

func fetchSubmissions(state *types.ApplicationState, cat *types.Category) ([]*types.Submission, error) {
	state.PageToFetchFromAPI++
	return fetcher.FetchSubmissions(state.PageToFetchFromAPI, cat.CurrentCategory)
}

func getListFromFrontPanel(pages *cview.Panels) *cview.List {
	_, primitive := pages.GetFrontPanel()
	list, _ := primitive.(*cview.List)
	return list
}

func setShortcuts(app *cview.Application,
	state []*types.ApplicationState,
	main *primitives.MainView,
	cat *types.Category) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentState := state[cat.CurrentCategory]
		currentPage := currentState.CurrentPage
		screenWidth := currentState.ScreenWidth
		viewableStories := currentState.ViewableStoriesOnSinglePage

		frontPanel, _ := main.Panels.GetFrontPanel()

		if frontPanel == offlinePage {
			if event.Rune() == 'q' {
				app.Stop()
			}
			return event
		}

		if frontPanel == helpPage {
			view.SetHackerNewsHeader(main, screenWidth, cat.CurrentCategory)
			view.SetPanelCategory(main, cat.CurrentCategory)
			view.SetFooterText(main, currentPage, screenWidth, currentState.MaxPages)
			view.SetLeftMarginRanks(main, currentPage, viewableStories)
			return event
		}

		if event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab {
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
			setList(list, nextState.Submissions, 0, nextState.ViewableStoriesOnSinglePage, app)

			view.SetFooterText(main, nextState.CurrentPage, nextState.ScreenWidth, nextState.MaxPages)
			view.SetLeftMarginRanks(main, nextState.CurrentPage, nextState.ViewableStoriesOnSinglePage)
			view.SetHackerNewsHeader(main, nextState.ScreenWidth, cat.CurrentCategory)

			return event
		}

		if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			nextPage(app, currentState, main, cat)
		} else if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			previousPage(app, currentState, main, main.Panels)
		} else if event.Rune() == 'q' || event.Key() == tcell.KeyEsc {
			app.Stop()
		} else if event.Rune() == 'i' || event.Rune() == '?' {
			view.SetKeymapsHeader(main, screenWidth)
			view.HideLeftMarginRanks(main)
			view.HideFooterText(main)
			view.SetPanelToHelpScreen(main)
		} else if event.Rune() == 'g' {
			view.SelectFirstElementInList(main)
		} else if event.Rune() == 'G' {
			view.SelectLastElementInList(currentState, main)
		}
		return event
	})
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

func nextPage(app *cview.Application, state *types.ApplicationState, main *primitives.MainView, cat *types.Category) {
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

func pageHasEnoughSubmissionsToView(page int, visibleStories int, submissions []*types.Submission) bool {
	largestItemToDisplay := (page * visibleStories) + visibleStories
	downloadedSubmissions := len(submissions)

	return downloadedSubmissions > largestItemToDisplay
}

func getCurrentlySelectedItemOnFrontPage(pages *cview.Panels) int {
	_, primitive := pages.GetFrontPanel()
	list, ok := primitive.(*cview.List)
	if ok {
		return list.GetCurrentItemIndex()
	}
	return 0
}

func previousPage(app *cview.Application,
	state *types.ApplicationState,
	main *primitives.MainView,
	pages *cview.Panels) {

	previousPage := state.CurrentPage - 1
	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(pages)

	if previousPage < 0 {
		return
	}

	list := getListFromFrontPanel(pages)

	setList(list, state.Submissions, previousPage, state.ViewableStoriesOnSinglePage, app)
	list.SetCurrentItem(currentlySelectedItem)

	state.CurrentPage--

	view.SetLeftMarginRanks(main, state.CurrentPage, state.ViewableStoriesOnSinglePage)
	view.SetFooterText(main, state.CurrentPage, state.ScreenWidth, state.MaxPages)
}

func setSelectedFunction(
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

func createNewList() *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)

	return list
}
