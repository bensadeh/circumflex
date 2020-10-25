package controller

import (
	"clx/browser"
	"clx/cli"
	cp "clx/comment-parser"
	"clx/primitives"
	"clx/screen"
	formatter2 "clx/submission/formatter"
	"clx/types"
	"encoding/json"
	text "github.com/MichaelMure/go-term-text"
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
	Submissions                 []types.Submission
	MappedSubmissions           int
	MappedPages                 int
	StoriesListed               int
	Application                 *cview.Application
	PageToFetchFromAPI          int
	CurrentPage                 int
	ScreenHeight                int
	ScreenWidth                 int
	ViewableStoriesOnSinglePage int
	MaxPages                    int
	IsOffline                   bool
	MainView                    *primitives.MainView
}

func NewScreenController() *screenController {
	sc := new(screenController)
	sc.MaxPages = maxPages
	sc.Application = cview.NewApplication()
	sc.ScreenHeight = screen.GetTerminalHeight()
	sc.ScreenWidth = screen.GetTerminalWidth()
	sc.ViewableStoriesOnSinglePage = screen.GetViewableStoriesOnSinglePage(sc.ScreenHeight, maximumStoriesToDisplay)
	sc.MainView = primitives.NewMainView(sc.ScreenWidth, sc.ViewableStoriesOnSinglePage)
	sc.setShortcuts()
	submissions, err := sc.fetchSubmissions()
	sc.IsOffline = getIsOfflineStatus(err)

	sc.mapSubmissions(sc.Application, submissions, sc.CurrentPage, sc.ViewableStoriesOnSinglePage)

	startPage := getStartPage(sc.IsOffline)
	sc.MainView.Pages.SwitchToPage(startPage)

	return sc
}

func (sc *screenController) getHeadline() string {
	base := "[black:orange:]   [Y[] Hacker News"
	offset := -16
	whitespace := ""
	for i := 0; i < sc.ScreenWidth-text.Len(base)-offset; i++ {
		whitespace += " "
	}
	return base + whitespace
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
	return strconv.Itoa(sc.CurrentPage)
}

func (sc *screenController) setShortcuts() {
	app := sc.Application
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentPage, _ := sc.MainView.Pages.GetFrontPage()

		if currentPage == offlinePage {
			if event.Rune() == 'q' {
				app.Stop()
			}
			return event
		}

		if currentPage == helpPage {
			sc.MainView.SetHeaderTextToHN(sc.ScreenWidth)
			sc.MainView.Pages.SwitchToPage(sc.getCurrentPage())
			sc.MainView.SetFooterText(sc.CurrentPage, sc.ScreenWidth)
			sc.MainView.SetLeftMarginRanks(sc.CurrentPage, sc.ViewableStoriesOnSinglePage)
			return event
		}

		if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			sc.nextPage(sc.CurrentPage, sc.MaxPages, sc.MainView.Pages, sc.MappedPages, sc.Application)
			sc.MainView.SetLeftMarginRanks(sc.CurrentPage, sc.ViewableStoriesOnSinglePage)
		} else if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			sc.previousPage(sc.CurrentPage, sc.MainView.Pages)
			sc.MainView.SetLeftMarginRanks(sc.CurrentPage, sc.ViewableStoriesOnSinglePage)
		} else if event.Rune() == 'q' {
			app.Stop()
		} else if event.Rune() == 'i' || event.Rune() == '?' {
			sc.MainView.SetHeaderTextToKeymaps(sc.ScreenWidth)
			sc.MainView.HideFooterText()
			sc.MainView.HideLeftMarginRanks()
			sc.MainView.Pages.SwitchToPage(helpPage)
		}
		return event
	})
}

func (sc *screenController) nextPage(currentPage int, maxPages int, pages *cview.Pages, mappedPages int, app *cview.Application) {
	nextPage := currentPage + 1

	if nextPage > maxPages {
		return
	}

	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(pages)

	if nextPage < mappedPages {
		pages.SwitchToPage(strconv.Itoa(nextPage))
		app.ForceDraw()
		setCurrentlySelectedItemOnFrontPage(currentlySelectedItem, pages)
		app.ForceDraw()
	} else {
		submissions, _ := sc.fetchSubmissions()
		sc.mapSubmissions(sc.Application, submissions, sc.CurrentPage, sc.ViewableStoriesOnSinglePage)
		pages.SwitchToPage(strconv.Itoa(nextPage))

		app.ForceDraw()
		setCurrentlySelectedItemOnFrontPage(currentlySelectedItem, pages)
		app.ForceDraw()
	}

	sc.CurrentPage++
	sc.MainView.SetFooterText(sc.CurrentPage, sc.ScreenWidth)
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

func (sc *screenController) previousPage(currentPage int, pages *cview.Pages) {
	previousPage := currentPage - 1

	if previousPage < 0 {
		return
	}

	currentlySelectedItem := getCurrentlySelectedItemOnFrontPage(pages)

	sc.CurrentPage--
	pages.SwitchToPage(strconv.Itoa(sc.CurrentPage))

	setCurrentlySelectedItemOnFrontPage(currentlySelectedItem, pages)
	sc.MainView.SetFooterText(sc.CurrentPage, sc.ScreenWidth)
}

func (sc *screenController) getStoriesToDisplay() int {
	return sc.ViewableStoriesOnSinglePage
}

func setSelectedFunction(app *cview.Application, list *cview.List, submissions []types.Submission, currentPage int, viewableStoriesOnSinglePage int) {
	list.SetSelectedFunc(func(i int, _ *cview.ListItem) {
		app.Suspend(func() {
			for index := range submissions {
				if index == i {
					storyIndex := (currentPage)*viewableStoriesOnSinglePage + i
					s := submissions[storyIndex]
					id := strconv.Itoa(s.ID)
					JSON, _ := get("http://node-hnapi.herokuapp.com/item/" + id)
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
			item := list.GetCurrentItemIndex() + viewableStoriesOnSinglePage*(currentPage)
			url := submissions[item].URL
			browser.Open(url)
		}
		return event
	})
}

func (sc *screenController) getSubmission(i int) types.Submission {
	return sc.Submissions[i]
}

func (sc *screenController) fetchSubmissions() ([]types.Submission, error) {
	sc.PageToFetchFromAPI++
	p := strconv.Itoa(sc.PageToFetchFromAPI)
	return getSubmissions(p)
}

func (sc *screenController) mapSubmissions(app *cview.Application,  submissions []types.Submission, currentPage int, viewableStoriesOnSinglePage int) {
	sc.Submissions = append(sc.Submissions, submissions...)
	sc.mapSubmissionsToListItems(app, submissions, currentPage, viewableStoriesOnSinglePage)
}

func (sc *screenController) mapSubmissionsToListItems(app *cview.Application,  submissions []types.Submission, currentPage int, viewableStoriesOnSinglePage int) {
	for sc.hasStoriesToMap() {
		sub := sc.Submissions[sc.MappedSubmissions : sc.MappedSubmissions+sc.ViewableStoriesOnSinglePage]
		list := createNewList(app, submissions, currentPage, viewableStoriesOnSinglePage)
		addSubmissionsToList(list, sub, sc)

		sc.MainView.Pages.AddPage(strconv.Itoa(sc.MappedPages), list, true, true)
		sc.MappedPages++
	}
}

func (sc *screenController) hasStoriesToMap() bool {
	return len(sc.Submissions)-sc.MappedSubmissions >= sc.ViewableStoriesOnSinglePage
}

func createNewList(app *cview.Application,  submissions []types.Submission, currentPage int, viewableStoriesOnSinglePage int) *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)
	setSelectedFunction(app, list, submissions, currentPage, viewableStoriesOnSinglePage)

	return list
}

func addSubmissionsToList(list *cview.List, submissions []types.Submission, sh *screenController) {
	for _, s := range submissions {
		mainText := formatter2.GetMainText(s.Title, s.Domain)
		secondaryText := formatter2.GetSecondaryText(s.Points, s.Author, s.Time, s.CommentsCount)

		item := cview.NewListItem(mainText)
		item.SetSecondaryText(secondaryText)

		list.AddItem(item)
		sh.MappedSubmissions++
	}
}
