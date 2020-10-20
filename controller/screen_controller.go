package controller

import (
	"clx/browser"
	"clx/cli"
	commentparser "clx/comment-parser"
	"clx/model"
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
	Pages                       *cview.Pages
	Application                 *cview.Application
	PageToFetchFromAPI          int
	CurrentPage                 int
	ScreenHeight                int
	ScreenWidth                 int
	ViewableStoriesOnSinglePage int
	MaxPages                    int
	IsOffline                   bool
	Grid                        *cview.Grid
	Footer                      *cview.TextView
}

func NewScreenController() *screenController {
	sc := new(screenController)
	sc.Application = cview.NewApplication()
	sc.setShortcuts()
	sc.Pages = cview.NewPages()
	sc.MaxPages = maxPages
	sc.ScreenHeight = model.GetTerminalHeight()
	sc.ScreenWidth = model.GetTerminalWidth()
	sc.ViewableStoriesOnSinglePage = model.GetViewableStoriesOnSinglePage(sc.ScreenHeight, maximumStoriesToDisplay)
	submissions, err := sc.fetchSubmissions()
	sc.IsOffline = getIsOfflineStatus(err)
	sc.mapSubmissions(submissions)

	newPrimitive := func(text string) cview.Primitive {
		tv := cview.NewTextView()
		tv.SetTextAlign(cview.AlignLeft)
		tv.SetText(text)
		tv.SetBorder(false)
		tv.SetBackgroundColor(tcell.ColorDefault)
		tv.SetTextColor(tcell.ColorDefault)
		tv.SetDynamicColors(true)
		return tv
	}
	leftMargin := newPrimitive("")
	rightMargin := newPrimitive("")
	main := sc.Pages

	grid := cview.NewGrid()
	grid.SetBorder(false)
	grid.SetRows(2, 0, 1)
	grid.SetColumns(3, 0, 3)
	grid.SetBackgroundColor(tcell.ColorDefault)
	grid.AddItem(newPrimitive(sc.getHeadline()), 0, 0, 1, 3, 0, 0, false)
	sc.Footer = newPrimitive(sc.getFooterText()).(*cview.TextView)
	grid.AddItem(sc.Footer, 2, 0, 1, 3, 0, 0, false)

	grid.AddItem(leftMargin, 1, 0, 1, 1, 0, 0, false)
	grid.AddItem(main, 1, 1, 1, 1, 0, 0, true)
	grid.AddItem(rightMargin, 1, 2, 1, 1, 0, 0, false)

	sc.Grid = grid

	sc.Pages.AddPage(helpPage, getHelpScreen(), true, false)
	sc.Pages.AddPage(offlinePage, getOfflineScreen(), true, false)

	startPage := getStartPage(sc.IsOffline)
	sc.Pages.SwitchToPage(startPage)

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

func (sc *screenController) getFooterText() string {
	page := ""
	switch sc.CurrentPage {
	case 0:
		page = "   •◦◦"
	case 1:
		page = "   ◦•◦"
	case 2:
		page = "   ◦◦•"
	default:
		page = ""
	}
	return sc.rightPadWithWhitespace(page)
}

func (sc *screenController) rightPadWithWhitespace(s string) string {
	offset := 3
	whitespace := ""
	for i := 0; i < sc.ScreenWidth-text.Len(s)-offset; i++ {
		whitespace += " "
	}
	return whitespace + s
}

func (sc *screenController) setShortcuts() {
	app := sc.Application
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentPage, _ := sc.Pages.GetFrontPage()

		if currentPage == offlinePage {
			if event.Rune() == 'q' {
				app.Stop()
			}
			return event
		}

		if currentPage == helpPage {
			sc.Pages.SwitchToPage(sc.getCurrentPage())
			return event
		}

		if event.Rune() == 'l' || event.Key() == tcell.KeyRight {
			sc.nextPage()
		} else if event.Rune() == 'h' || event.Key() == tcell.KeyLeft {
			sc.previousPage()
		} else if event.Rune() == 'q' {
			app.Stop()
		} else if event.Rune() == 'i' || event.Rune() == '?' {
			sc.Pages.SwitchToPage(helpPage)
		}
		return event
	})
}

func (sc *screenController) nextPage() {
	nextPage := sc.CurrentPage + 1

	if nextPage > sc.MaxPages {
		return
	}

	_, primitive := sc.Pages.GetFrontPage()
	list := primitive.(*cview.List)
	currentlySelectedItem := list.GetCurrentItemIndex()

	if nextPage < sc.MappedPages {
		sc.Pages.SwitchToPage(strconv.Itoa(nextPage))
		_, p := sc.Pages.GetFrontPage()
		l := p.(*cview.List)
		sc.Application.ForceDraw()
		l.SetCurrentItem(currentlySelectedItem)
		sc.Application.ForceDraw()
	} else {
		submissions, _ := sc.fetchSubmissions()
		sc.mapSubmissions(submissions)
		sc.Pages.SwitchToPage(strconv.Itoa(nextPage))

		_, p := sc.Pages.GetFrontPage()
		l := p.(*cview.List)
		sc.Application.ForceDraw()
		l.SetCurrentItem(currentlySelectedItem)
		sc.Application.ForceDraw()
	}

	sc.CurrentPage++
	sc.Footer.SetText(sc.getFooterText())
}

func (sc *screenController) previousPage() {
	previousPage := sc.CurrentPage - 1

	if previousPage < 0 {
		return
	}

	_, primitive := sc.Pages.GetFrontPage()
	list := primitive.(*cview.List)
	currentlySelectedItem := list.GetCurrentItemIndex()

	sc.CurrentPage--
	sc.Pages.SwitchToPage(strconv.Itoa(sc.CurrentPage))

	_, p := sc.Pages.GetFrontPage()
	l := p.(*cview.List)
	l.SetCurrentItem(currentlySelectedItem)
	sc.Footer.SetText(sc.getFooterText())
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
					jComments := new(commentparser.Comments)
					_ = json.Unmarshal(JSON, jComments)

					commentTree := commentparser.PrintCommentTree(*jComments, 4, 70)
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

		sc.Pages.AddPage(strconv.Itoa(sc.MappedPages), list, true, true)
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
		mainText := model.GetMainText(s.Title, s.Domain, sh.MappedSubmissions)
		secondaryText := model.GetSecondaryText(s.Points, s.Author, s.Time, s.CommentsCount)

		item := cview.NewListItem(mainText)
		item.SetSecondaryText(secondaryText)

		list.AddItem(item)
		sh.MappedSubmissions++
	}
}
