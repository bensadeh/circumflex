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
	"unicode"
)

const (
	helpPage    = "help"
	offlinePage = "offline"
)

func SetApplicationShortcuts(app *cview.Application,
	submissionStates []*types.SubmissionState,
	main *types.MainView,
	appState *types.ApplicationState) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentState := submissionStates[appState.CurrentCategory]

		frontPanel, _ := main.Panels.GetFrontPanel()

		if frontPanel == offlinePage {
			if event.Rune() == 'q' {
				app.Stop()
			}
			return event
		}

		if frontPanel == helpPage {
			ReturnFromHelpScreen(main, appState, currentState)
			return event
		}

		if event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab {
			ChangeCategory(event, appState, submissionStates, main, app)
			return event
		} else if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			NextPage(app, currentState, main, appState)
			return event
		} else if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			PreviousPage(app, currentState, main, appState)
			return event
		} else if event.Rune() == 'q' || event.Key() == tcell.KeyEsc {
			app.Stop()
		} else if event.Rune() == 'i' || event.Rune() == '?' {
			ShowHelpScreen(main, appState)
			return event
		} else if event.Rune() == 'g' {
			SelectFirstElementInList(main)
			return event
		} else if event.Rune() == 'G' {
			SelectLastElementInList(main, appState)
			return event
		} else if event.Rune() == 'r' {
			ResetStates(appState, submissionStates)
			InitializeHeaderAndFooterAndLeftMarginView(appState, submissionStates, main)
			FetchAndAppendSubmissions(submissionStates[appState.CurrentCategory], appState)
			ShowPageAfterResize(appState, submissionStates, main, app)
			return event
		} else if unicode.IsDigit(event.Rune()) {
			SelectElementInList(main, event.Rune())
			return event
		}
		return event
	})
}

func SetListShortcuts(
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

func NextPage(app *cview.Application,
	subState *types.SubmissionState,
	main *types.MainView,
	appState *types.ApplicationState) {

	nextPage := appState.CurrentPage + 1

	if nextPage > subState.MaxPages {
		return
	}

	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(main.Panels)

	list := GetListFromFrontPanel(main.Panels)

	if !pageHasEnoughSubmissionsToView(nextPage, appState.ViewableStoriesOnSinglePage, subState.Submissions) {
		FetchAndAppendSubmissions(subState, appState)
	}

	appState.CurrentPage++

	ShowSubmissions(list, subState.Submissions, appState, app)
	list.SetCurrentItem(currentlySelectedItem)

	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.ViewableStoriesOnSinglePage)
	view.SetFooter(main, appState.CurrentPage, appState.ScreenWidth, subState.MaxPages)
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

func FetchAndAppendSubmissions(state *types.SubmissionState, cat *types.ApplicationState) {
	newSubs, _ := FetchSubmissions(state, cat)
	state.Submissions = append(state.Submissions, newSubs...)
}

func FetchSubmissions(state *types.SubmissionState, cat *types.ApplicationState) ([]*types.Submission, error) {
	state.PageToFetchFromAPI++
	return fetcher.FetchSubmissions(state.PageToFetchFromAPI, cat.CurrentCategory)
}

func ShowSubmissions(list *cview.List, submissions []*types.Submission, appState *types.ApplicationState, app *cview.Application) {
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

	SetListShortcuts(app, list, submissions, appState)
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
	appState.CurrentPage = 0

	if !pageHasEnoughSubmissionsToView(0, appState.ViewableStoriesOnSinglePage, nextState.Submissions) {
		FetchAndAppendSubmissions(nextState, appState)
	}

	view.SetPanelCategory(main, appState.CurrentCategory)
	list := GetListFromFrontPanel(main.Panels)
	ShowSubmissions(list, nextState.Submissions, appState, app)

	view.SetFooter(main, appState.CurrentPage, appState.ScreenWidth, nextState.MaxPages)
	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.ViewableStoriesOnSinglePage)
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

	previousPage := appState.CurrentPage - 1
	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(main.Panels)

	if previousPage < 0 {
		return
	}

	list := GetListFromFrontPanel(main.Panels)

	appState.CurrentPage--

	ShowSubmissions(list, state.Submissions, appState, app)
	list.SetCurrentItem(currentlySelectedItem)

	view.SetLeftMarginRanks(main, appState.CurrentPage, appState.ViewableStoriesOnSinglePage)
	view.SetFooter(main, appState.CurrentPage, appState.ScreenWidth, state.MaxPages)
}

func ShowHelpScreen(main *types.MainView, appState *types.ApplicationState) {
	appState.IsOnHelpScreen = true

	view.SetKeymapsHeader(main, appState.ScreenWidth)
	view.HideLeftMarginRanks(main)
	view.HideFooterText(main)
	view.SetPanelToHelpScreen(main)
}

func ReturnFromHelpScreen(main *types.MainView, appState *types.ApplicationState, subState *types.SubmissionState) {
	appState.IsOnHelpScreen = false

	view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.CurrentCategory)
	view.SetPanelCategory(main, appState.CurrentCategory)
	view.SetFooter(main, appState.CurrentPage, appState.ScreenWidth, subState.MaxPages)
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

func ResetStates(appState *types.ApplicationState, submissionStates []*types.SubmissionState) {
	resetApplicationState(appState)
	resetSubmissionStates(submissionStates)
}

func resetApplicationState(appState *types.ApplicationState) {
	appState.CurrentPage = 0
	appState.ScreenWidth = screen.GetTerminalWidth()
	appState.ScreenHeight = screen.GetTerminalHeight()
	appState.ViewableStoriesOnSinglePage = screen.GetViewableStoriesOnSinglePage(
		appState.ScreenHeight,
		30)
}

func resetSubmissionStates(submissionStates []*types.SubmissionState) {
	numberOfCategories := 3

	for i := 0; i < numberOfCategories; i++ {
		submissionStates[i].MappedSubmissions = 0
		submissionStates[i].PageToFetchFromAPI = 0
		submissionStates[i].StoriesListed = 0
		submissionStates[i].Submissions = nil
	}
}

func InitializeHeaderAndFooterAndLeftMarginView(appState *types.ApplicationState, submissionStates []*types.SubmissionState, main *types.MainView) {
	view.SetPanelCategory(main, appState.CurrentCategory)
	view.SetHackerNewsHeader(main, appState.ScreenWidth, appState.CurrentCategory)
	view.SetLeftMarginRanks(main, 0, appState.ViewableStoriesOnSinglePage)
	view.SetFooter(main,
		0,
		appState.ScreenWidth,
		submissionStates[appState.CurrentCategory].MaxPages)
}

func ShowPageAfterResize(appState *types.ApplicationState, submissionStates []*types.SubmissionState, main *types.MainView, app *cview.Application) {
	frontPanelList := GetListFromFrontPanel(main.Panels)

	ShowSubmissions(frontPanelList,
		submissionStates[appState.CurrentCategory].Submissions,
		appState,
		app)

	if appState.IsOnHelpScreen {
		ShowHelpScreen(main, appState)
	}

	SetApplicationShortcuts(app, submissionStates, main, appState)
}
