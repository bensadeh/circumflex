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

	sc.ApplicationState[types.NoCategory].MaxPages = 2
	sc.ApplicationState[types.NoCategory].ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState[types.NoCategory].ScreenHeight = screen.GetTerminalHeight()
	sc.ApplicationState[types.NoCategory].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.ApplicationState[types.New].MaxPages = 2
	sc.ApplicationState[types.New].ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState[types.New].ScreenHeight = screen.GetTerminalHeight()
	sc.ApplicationState[types.New].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.ApplicationState[types.Ask].MaxPages = 1
	sc.ApplicationState[types.Ask].ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState[types.Ask].ScreenHeight = screen.GetTerminalHeight()
	sc.ApplicationState[types.Ask].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.ApplicationState[types.Show].MaxPages = 1
	sc.ApplicationState[types.Show].ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState[types.Show].ScreenHeight = screen.GetTerminalHeight()
	sc.ApplicationState[types.Show].ViewableStoriesOnSinglePage = storiesToDisplay

	sc.MainView = primitives.NewMainView(
		sc.ApplicationState[types.NoCategory].ScreenWidth,
		sc.ApplicationState[types.NoCategory].ViewableStoriesOnSinglePage)

	newSubmissions, _ := fetchSubmissions(sc.ApplicationState[types.NoCategory], sc.Category)

	mapSubmissions(sc.Application,
		sc.ApplicationState,
		newSubmissions,
		sc.MainView,
		sc.Category)

	sc.MainView.Panels.SetCurrentPanel("0-0")
	_, primitive := sc.MainView.Panels.GetFrontPanel()
	list, ok := primitive.(*cview.List)

	if ok {
		setList(list, newSubmissions, 0, storiesToDisplay)
	}

	setShortcuts(sc.Application,
		sc.ApplicationState,
		sc.MainView,
		sc.Category)

	return sc
}

func setList(list *cview.List, submissions []*types.Submission, page int, submissionsToShow int) {
	list.Clear()
	start := page * submissionsToShow
	end := start + submissionsToShow

	for i := start; i < end; i++ {
		s := submissions[i]
		mainText := formatter2.GetMainText(s.Title, s.Domain)
		secondaryText := formatter2.GetSecondaryText(s.Points, s.Author, s.Time, s.CommentsCount)

		item := cview.NewListItem(mainText + "!")
		item.SetSecondaryText(secondaryText + "!")

		list.AddItem(item)
	}
}

func fetchAndAppendSubmissions(state *types.ApplicationState, cat *types.Category) {
	newSubs, _ := fetchSubmissions(state, cat)
	state.Submissions = append(state.Submissions, newSubs...)
}

func fetchSubmissions(state *types.ApplicationState, cat *types.Category) ([]*types.Submission, error) {
	state.PageToFetchFromAPI++
	return fetcher.FetchSubmissions(state.PageToFetchFromAPI, cat.CurrentCategory)
}

func getPage(currentPage int, currentCategory int) string {
	return strconv.Itoa(currentPage) + "-" + strconv.Itoa(currentCategory)
}

func setShortcuts(app *cview.Application,
	state []*types.ApplicationState,
	main *primitives.MainView,
	cat *types.Category) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentState := state[cat.CurrentCategory]

		frontPanel, _ := main.Panels.GetFrontPanel()

		if frontPanel == offlinePage {
			if event.Rune() == 'q' {
				app.Stop()
			}
			return event
		}

		if frontPanel == helpPage {
			main.SetHeaderTextCategory(currentState.ScreenWidth, cat.CurrentCategory)
			page := getPage(currentState.CurrentPage, cat.CurrentCategory)
			main.Panels.SetCurrentPanel(page)
			main.SetFooterText(currentState.CurrentPage, currentState.ScreenWidth, currentState.MaxPages)
			main.SetLeftMarginRanks(currentState.CurrentPage, currentState.ViewableStoriesOnSinglePage)
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

			if len(state[cat.CurrentCategory].Submissions) == 0 {
				newSubmissions, _ := fetchSubmissions(currentState, cat)
				mapSubmissions(app,
					state,
					newSubmissions,
					main,
					cat)
			}

			pageToView := getPage(0, cat.CurrentCategory)
			main.Panels.SetCurrentPanel(pageToView)
			main.SetFooterText(nextState.CurrentPage, nextState.ScreenWidth, nextState.MaxPages)
			main.SetLeftMarginRanks(nextState.CurrentPage, nextState.ViewableStoriesOnSinglePage)
			main.SetHeaderTextCategory(nextState.ScreenWidth, cat.CurrentCategory)
			setCurrentlySelectedItemOnFrontPage(0, main.Panels)

			return event
		}

		if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			nextPage(app, state, main, cat)
			main.SetLeftMarginRanks(currentState.CurrentPage,
				currentState.ViewableStoriesOnSinglePage)
			main.SetFooterText(currentState.CurrentPage,
				currentState.ScreenWidth, currentState.MaxPages)
		} else if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			previousPage(currentState, main.Panels, cat)
			main.SetLeftMarginRanks(currentState.CurrentPage,
				currentState.ViewableStoriesOnSinglePage)
			main.SetFooterText(currentState.CurrentPage,
				currentState.ScreenWidth, currentState.MaxPages)
		} else if event.Rune() == 'q' || event.Key() == tcell.KeyEsc {
			app.Stop()
		} else if event.Rune() == 'i' || event.Rune() == '?' {
			main.SetHeaderTextToKeymaps(currentState.ScreenWidth)
			main.HideFooterText()
			main.HideLeftMarginRanks()
			main.Panels.SetCurrentPanel(helpPage)
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

func nextPage(app *cview.Application,
	state []*types.ApplicationState,
	main *primitives.MainView,
	cat *types.Category) {
	currentState := state[cat.CurrentCategory]
	nextPage := currentState.CurrentPage + 1

	if nextPage > currentState.MaxPages {
		return
	}

	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(main.Panels)

	_, primitive := main.Panels.GetFrontPanel()
	list, _ := primitive.(*cview.List)

	if !pageHasEnoughSubmissionsToView(nextPage, currentState.ViewableStoriesOnSinglePage, currentState.Submissions) {
		fetchAndAppendSubmissions(currentState, cat)
	}

	setList(list, currentState.Submissions, nextPage, currentState.ViewableStoriesOnSinglePage)
	list.SetCurrentItem(currentlySelectedItem)

	currentState.CurrentPage++
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

func setCurrentlySelectedItemOnFrontPage(item int, pages *cview.Panels) {
	_, primitive := pages.GetFrontPanel()
	list, ok := primitive.(*cview.List)
	if ok {
		list.SetCurrentItem(item)
	}
}

func previousPage(state *types.ApplicationState, pages *cview.Panels, cat *types.Category) {
	previousPage := state.CurrentPage - 1
	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(pages)

	if previousPage < 0 {
		return
	}

	_, primitive := pages.GetFrontPanel()
	list, _ := primitive.(*cview.List)

	setList(list, state.Submissions, previousPage, state.ViewableStoriesOnSinglePage)
	list.SetCurrentItem(currentlySelectedItem)

	state.CurrentPage--
}

func setSelectedFunction(app *cview.Application,
	list *cview.List,
	state []*types.ApplicationState,
	cat *types.Category) {
	currentState := state[cat.CurrentCategory]

	list.SetSelectedFunc(func(i int, _ *cview.ListItem) {
		app.Suspend(func() {
			for index := range currentState.Submissions {
				if index == i {
					storyIndex := (currentState.CurrentPage)*currentState.ViewableStoriesOnSinglePage + i
					s := currentState.Submissions[storyIndex]

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
			item := list.GetCurrentItemIndex() + currentState.ViewableStoriesOnSinglePage*(currentState.CurrentPage)
			url := currentState.Submissions[item].URL
			browser.Open(url)
		}
		if event.Key() == tcell.KeyTAB {

			return event
		}
		if event.Key() == tcell.KeyTab {

			return event
		}
		return event
	})
}

func mapSubmissions(app *cview.Application,
	state []*types.ApplicationState,
	newSubmissions []*types.Submission,
	main *primitives.MainView,
	cat *types.Category) {
	currentState := state[cat.CurrentCategory]
	currentState.Submissions = append(currentState.Submissions, newSubmissions...)
	mapSubmissionsToListItems(app, state, main, cat)
}

func mapSubmissionsToListItems(app *cview.Application,
	state []*types.ApplicationState,
	main *primitives.MainView,
	cat *types.Category) {
	currentState := state[cat.CurrentCategory]

	for hasStoriesToMap(currentState.Submissions, currentState) {
		sub := currentState.Submissions[currentState.MappedSubmissions : currentState.MappedSubmissions+currentState.ViewableStoriesOnSinglePage]
		list := createNewList(app, state, cat)
		addSubmissionsToList(list, sub, currentState)

		pageName := getPage(currentState.MappedPages, cat.CurrentCategory)
		main.Panels.AddPanel(pageName, list, true, false)
		currentState.MappedPages++
	}
}

func hasStoriesToMap(submissions []*types.Submission, state *types.ApplicationState) bool {
	return len(submissions)-state.MappedSubmissions >= state.ViewableStoriesOnSinglePage
}

func createNewList(app *cview.Application,
	state []*types.ApplicationState,
	cat *types.Category) *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)
	setSelectedFunction(app, list, state, cat)

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
