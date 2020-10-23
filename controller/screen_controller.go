package controller

import (
	"clx/browser"
	"clx/cli"
	parser "clx/comment-parser"
	"clx/primitives"
	"clx/screen"
	"clx/submission-parser"
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
	maxPages                = 2
)

type screenController struct {
	Submissions                 []Submission
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
	sc.MainView = primitives.NewMainView(sc.ScreenWidth)
	sc.setShortcuts()
	sc.ViewableStoriesOnSinglePage = screen.GetViewableStoriesOnSinglePage(sc.ScreenHeight, maximumStoriesToDisplay)
	submissions, err := sc.fetchSubmissions()
	sc.IsOffline = getIsOfflineStatus(err)
	sc.mapSubmissions(submissions)

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

//func (sc *screenController) getFooterText() string {
//	page := ""
//	switch sc.CurrentPage {
//	case 0:
//		page = "   •◦◦"
//	case 1:
//		page = "   ◦•◦"
//	case 2:
//		page = "   ◦◦•"
//	default:
//		page = ""
//	}
//	return sc.rightPadWithWhitespace(page)
//}
//
//func (sc *screenController) rightPadWithWhitespace(s string) string {
//	offset := 3
//	whitespace := ""
//	for i := 0; i < sc.ScreenWidth-text.Len(s)-offset; i++ {
//		whitespace += " "
//	}
//	return whitespace + s
//}

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
			sc.MainView.Pages.SwitchToPage(sc.getCurrentPage())
			return event
		}

		if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			sc.nextPage(sc.CurrentPage, sc.MaxPages, sc.MainView.Pages, sc.MappedPages, sc.Application)
		} else if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			sc.previousPage()
		} else if event.Rune() == 'q' {
			app.Stop()
		} else if event.Rune() == 'i' || event.Rune() == '?' {
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
		sc.mapSubmissions(submissions)
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

func (sc *screenController) previousPage() {
	previousPage := sc.CurrentPage - 1

	if previousPage < 0 {
		return
	}

	_, primitive := sc.MainView.Pages.GetFrontPage()
	list := primitive.(*cview.List)
	currentlySelectedItem := list.GetCurrentItemIndex()

	sc.CurrentPage--
	sc.MainView.Pages.SwitchToPage(strconv.Itoa(sc.CurrentPage))

	_, p := sc.MainView.Pages.GetFrontPage()
	l := p.(*cview.List)
	l.SetCurrentItem(currentlySelectedItem)
	sc.MainView.SetFooterText(sc.CurrentPage, sc.ScreenWidth)
}

func (sc *screenController) getStoriesToDisplay() int {
	return sc.ViewableStoriesOnSinglePage
}

func setSelectedFunction(app *cview.Application, list *cview.List, sh *screenController) {
	list.SetSelectedFunc(func(i int, _ *cview.ListItem) {
		app.Suspend(func() {
			for index := range sh.Submissions {
				if index == i {
					id := getSubmissionID(i, sh)
					JSON, _ := get("http://node-hnapi.herokuapp.com/item/" + id)
					jComments := new(parser.Comments)
					_ = json.Unmarshal(JSON, jComments)

					commentTree := parser.PrintCommentTree(*jComments, 4, 70)
					cli.Less(commentTree)
				}
			}
		})
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'o' {
			item := list.GetCurrentItemIndex()
			url := sh.Submissions[item].URL
			browser.Open(url)
		}
		return event
	})
}

func getSubmissionID(i int, sh *screenController) string {
	storyIndex := (sh.CurrentPage)*sh.ViewableStoriesOnSinglePage + i
	s := sh.Submissions[storyIndex]
	return strconv.Itoa(s.ID)
}

func (sc *screenController) getSubmission(i int) Submission {
	return sc.Submissions[i]
}

type Submission struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Points        int    `json:"points"`
	Author        string `json:"user"`
	Time          string `json:"time_ago"`
	CommentsCount int    `json:"comments_count"`
	URL           string `json:"url"`
	Domain        string `json:"domain"`
	Type          string `json:"type"`
}

func (sc *screenController) fetchSubmissions() ([]Submission, error) {
	sc.PageToFetchFromAPI++
	p := strconv.Itoa(sc.PageToFetchFromAPI)
	return getSubmissions(p)
}

func (sc *screenController) mapSubmissions(submissions []Submission) {
	sc.Submissions = append(sc.Submissions, submissions...)
	sc.mapSubmissionsToListItems()
}

func (sc *screenController) mapSubmissionsToListItems() {
	for sc.hasStoriesToMap() {
		sub := sc.Submissions[sc.MappedSubmissions : sc.MappedSubmissions+sc.ViewableStoriesOnSinglePage]
		list := createNewList(sc)
		addSubmissionsToList(list, sub, sc)

		sc.MainView.Pages.AddPage(strconv.Itoa(sc.MappedPages), list, true, true)
		sc.MappedPages++
	}
}

func (sc *screenController) hasStoriesToMap() bool {
	return len(sc.Submissions)-sc.MappedSubmissions >= sc.ViewableStoriesOnSinglePage
}

func createNewList(sh *screenController) *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)
	setSelectedFunction(sh.Application, list, sh)

	return list
}

func addSubmissionsToList(list *cview.List, submissions []Submission, sh *screenController) {
	for _, s := range submissions {
		mainText := submission_parser.GetMainText(s.Title, s.Domain, sh.MappedSubmissions)
		secondaryText := submission_parser.GetSecondaryText(s.Points, s.Author, s.Time, s.CommentsCount)

		item := cview.NewListItem(mainText)
		item.SetSecondaryText(secondaryText)

		list.AddItem(item)
		sh.MappedSubmissions++
	}
}
