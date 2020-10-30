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

	sc.ApplicationState[types.NoCategory].MaxPages = 2
	sc.ApplicationState[types.NoCategory].ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState[types.NoCategory].ScreenHeight = screen.GetTerminalHeight()
	sc.ApplicationState[types.NoCategory].ViewableStoriesOnSinglePage = screen.GetViewableStoriesOnSinglePage(
		sc.ApplicationState[types.NoCategory].ScreenHeight,
		maximumStoriesToDisplay)

	sc.ApplicationState[types.New].MaxPages = 2
	sc.ApplicationState[types.New].ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState[types.New].ScreenHeight = screen.GetTerminalHeight()
	sc.ApplicationState[types.New].ViewableStoriesOnSinglePage = screen.GetViewableStoriesOnSinglePage(
		sc.ApplicationState[types.New].ScreenHeight,
		maximumStoriesToDisplay)

	sc.ApplicationState[types.Ask].MaxPages = 1
	sc.ApplicationState[types.Ask].ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState[types.Ask].ScreenHeight = screen.GetTerminalHeight()
	sc.ApplicationState[types.Ask].ViewableStoriesOnSinglePage = screen.GetViewableStoriesOnSinglePage(
		sc.ApplicationState[types.Ask].ScreenHeight,
		maximumStoriesToDisplay)

	sc.ApplicationState[types.Show].MaxPages = 1
	sc.ApplicationState[types.Show].ScreenWidth = screen.GetTerminalWidth()
	sc.ApplicationState[types.Show].ScreenHeight = screen.GetTerminalHeight()
	sc.ApplicationState[types.Show].ViewableStoriesOnSinglePage = screen.GetViewableStoriesOnSinglePage(
		sc.ApplicationState[types.Show].ScreenHeight,
		maximumStoriesToDisplay)

	sc.MainView = primitives.NewMainView(
		sc.ApplicationState[types.NoCategory].ScreenWidth,
		sc.ApplicationState[types.NoCategory].ViewableStoriesOnSinglePage)

	newSubmissions, err := fetchSubmissions(sc.ApplicationState[types.NoCategory], sc.Category)
	sc.ApplicationState[types.NoCategory].IsOffline = getIsOfflineStatus(err)

	mapSubmissions(sc.Application,
		sc.ApplicationState,
		newSubmissions,
		sc.MainView,
		sc.Category)

	startPage := getStartPage(sc.ApplicationState[types.NoCategory].IsOffline)
	sc.MainView.Pages.SwitchToPage(startPage)

	setShortcuts(sc.Application,
		sc.ApplicationState,
		sc.MainView,
		sc.Category)

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
	return "0-0"
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

		currentPage, _ := main.Pages.GetFrontPage()

		if currentPage == offlinePage {
			if event.Rune() == 'q' {
				app.Stop()
			}
			return event
		}

		if currentPage == helpPage {
			main.SetHeaderTextToHN(currentState.ScreenWidth)
			page := getPage(currentState.CurrentPage, cat.CurrentCategory)
			main.Pages.SwitchToPage(page)
			main.SetFooterText(currentState.CurrentPage, currentState.ScreenWidth, currentState.MaxPages)
			main.SetLeftMarginRanks(currentState.CurrentPage, currentState.ViewableStoriesOnSinglePage)
			return event
		}

		if event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab {
			isMovingBackwards := event.Key() == tcell.KeyBacktab
			cat.CurrentCategory = getNextCategory(cat.CurrentCategory, isMovingBackwards)
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
			main.Pages.SwitchToPage(pageToView)
			main.SetFooterText(nextState.CurrentPage, nextState.ScreenWidth, nextState.MaxPages)
			main.SetLeftMarginRanks(nextState.CurrentPage, nextState.ViewableStoriesOnSinglePage)
			main.SetHeaderTextCategory(nextState.ScreenWidth, cat.CurrentCategory)
			setCurrentlySelectedItemOnFrontPage(0, main.Pages)
			app.ForceDraw()

			return event
		}

		if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			nextPage(app, state, main, cat)
			main.SetLeftMarginRanks(currentState.CurrentPage,
				currentState.ViewableStoriesOnSinglePage)
			main.SetFooterText(currentState.CurrentPage,
				currentState.ScreenWidth, currentState.MaxPages)
		} else if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			previousPage(currentState, main.Pages, cat)
			main.SetLeftMarginRanks(currentState.CurrentPage,
				currentState.ViewableStoriesOnSinglePage)
			main.SetFooterText(currentState.CurrentPage,
				currentState.ScreenWidth, currentState.MaxPages)
		} else if event.Rune() == 'q' {
			app.Stop()
		} else if event.Rune() == 'i' || event.Rune() == '?' {
			main.SetHeaderTextToKeymaps(currentState.ScreenWidth)
			main.HideFooterText()
			main.HideLeftMarginRanks()
			main.Pages.SwitchToPage(helpPage)
		}
		return event
	})
}

func getNextCategory(currentCategory int, isMovingBackwards bool) int {
	if isMovingBackwards {
		return getPreviousCategory(currentCategory)
	}

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

	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(main.Pages)

	if nextPage < currentState.MappedPages {
		main.Pages.SwitchToPage(getPage(nextPage, cat.CurrentCategory))
		app.ForceDraw()
		setCurrentlySelectedItemOnFrontPage(currentlySelectedItem, main.Pages)
	} else {
		newSubmissions, _ := fetchSubmissions(currentState, cat)
		mapSubmissions(app,
			state,
			newSubmissions,
			main,
			cat)
		main.Pages.SwitchToPage(getPage(nextPage, cat.CurrentCategory))

		app.ForceDraw()
		setCurrentlySelectedItemOnFrontPage(currentlySelectedItem, main.Pages)
	}

	currentState.CurrentPage++

}

func getCurrentlySelectedItemOnFrontPage(pages *cview.Pages) int {
	_, primitive := pages.GetFrontPage()
	list, ok := primitive.(*cview.List)
	if ok {
		return list.GetCurrentItemIndex()
	}
	return 0
}

func setCurrentlySelectedItemOnFrontPage(item int, pages *cview.Pages) {
	_, primitive := pages.GetFrontPage()
	list, ok := primitive.(*cview.List)
	if ok {
		list.SetCurrentItem(item)
	}
}

func previousPage(state *types.ApplicationState, pages *cview.Pages, cat *types.Category) {
	previousPage := state.CurrentPage - 1

	if previousPage < 0 {
		return
	}

	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(pages)

	state.CurrentPage--
	pages.SwitchToPage(getPage(previousPage, cat.CurrentCategory))

	setCurrentlySelectedItemOnFrontPage(currentlySelectedItem, pages)
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

func fetchSubmissions(state *types.ApplicationState, cat *types.Category) ([]*types.Submission, error) {
	state.PageToFetchFromAPI++
	return fetcher.FetchSubmissions(state.PageToFetchFromAPI, cat.CurrentCategory)
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
		main.Pages.AddPage(pageName, list, true, true)
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
