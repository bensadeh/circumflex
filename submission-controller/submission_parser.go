package submission_controller

import (
	commentparser "clx/comment-parser"
	"clx/http-handler"
	http "clx/http-handler"
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"gitlab.com/tslocum/cview"
)

const (
	maximumStoriesToDisplay = 30
)

type submissionHandler struct {
	Submissions                 []submission
	MappedSubmissions           int
	MappedPages                 int
	StoriesListed               int
	Pages                       *cview.Pages
	Application                 *cview.Application
	PageToFetchFromAPI          int
	CurrentPage                 int
	ScreenHeight                int
	ViewableStoriesOnSinglePage int
	MaxPages                    int
}

func NewSubmissionHandler() *submissionHandler {
	sh := new(submissionHandler)
	sh.Application = cview.NewApplication()
	sh.setShortcuts()
	sh.Pages = cview.NewPages()
	sh.MaxPages = 2
	sh.ScreenHeight = getTerminalHeight()
	sh.ViewableStoriesOnSinglePage = min(sh.ScreenHeight/2, maximumStoriesToDisplay)
	sh.fetchSubmissions()

	sh.Pages.SwitchToPage("0")
	return sh
}

func (sh *submissionHandler) setShortcuts() {
	app := sh.Application
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			sh.nextPage()
		} else if event.Key() == tcell.KeyCtrlP {
			sh.previousPage()
		} else if event.Rune() == 'q' {
			app.Stop()
		}
		return event
	})
}

func getTerminalHeight() int {
	y, _ := terminal.Height()
	return int(y)
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (sh *submissionHandler) nextPage() {
	nextPage := sh.CurrentPage + 1

	if nextPage > sh.MaxPages {
		return
	}

	_, primitive := sh.Pages.GetFrontPage()
	list := primitive.(*cview.List)
	currentlySelectedItem := list.GetCurrentItem()

	if nextPage < sh.MappedPages {
		sh.Pages.SwitchToPage(strconv.Itoa(nextPage))
		_, p := sh.Pages.GetFrontPage()
		l := p.(*cview.List)
		l.SetCurrentItem(currentlySelectedItem)
	} else {
		sh.fetchSubmissions()
		sh.Pages.SwitchToPage(strconv.Itoa(nextPage))
	}

	sh.CurrentPage++
}

func (sh *submissionHandler) previousPage() {
	previousPage := sh.CurrentPage - 1

	if previousPage < 0 {
		return
	}

	_, primitive := sh.Pages.GetFrontPage()
	list := primitive.(*cview.List)
	currentlySelectedItem := list.GetCurrentItem()

	sh.CurrentPage--
	sh.Pages.SwitchToPage(strconv.Itoa(sh.CurrentPage))

	_, p := sh.Pages.GetFrontPage()
	l := p.(*cview.List)
	l.SetCurrentItem(currentlySelectedItem)
}

func (sh *submissionHandler) getStoriesToDisplay() int {
	return sh.ViewableStoriesOnSinglePage
}

func setSelectedFunction(app *cview.Application, list *cview.List, sh *submissionHandler) {
	list.SetSelectedFunc(func(i int, a string, b string, c rune) {
		app.Suspend(func() {
			for index := range sh.Submissions {
				if index == i {
					y, _ := terminal.Height()
					storiesToView := int(y / 2)
					storyRank := (sh.CurrentPage)*storiesToView + i

					id := strconv.Itoa(sh.Submissions[storyRank].ID)
					JSON, _ := http_handler.Get("http://node-hnapi.herokuapp.com/item/" + id)
					var jComments = new(commentparser.Comments)
					_ = json.Unmarshal(JSON, jComments)

					commentTree := commentparser.PrintCommentTree(*jComments, 4, 70)
					outputStringToLess(commentTree)
				}
			}
		})
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'o' {
			item := list.GetCurrentItem()
			url := sh.Submissions[item].URL
			openBrowser(url)
		}
		return event
	})
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func outputStringToLess(output string) {
	command := exec.Command("less", "-r")
	command.Stdin = strings.NewReader(output)
	command.Stdout = os.Stdout

	err := command.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func (sh *submissionHandler) getSubmission(i int) submission {
	return sh.Submissions[i]
}

type submission struct {
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

func (sh *submissionHandler) fetchSubmissions() {
	sh.PageToFetchFromAPI++
	p := strconv.Itoa(sh.PageToFetchFromAPI)
	JSON, _ := http.Get("http://node-hnapi.herokuapp.com/news?page=" + p)
	var submissions []submission
	_ = json.Unmarshal(JSON, &submissions)
	sh.Submissions = append(sh.Submissions, submissions...)
	sh.mapSubmissions()
}

func (sh *submissionHandler) mapSubmissions() {
	for sh.hasStoriesToMap() {
		sub := sh.Submissions[sh.MappedSubmissions : sh.MappedSubmissions+sh.ViewableStoriesOnSinglePage]
		list := createNewList(sh)
		addSubmissionsToList(list, sub, sh)

		sh.Pages.AddPage(strconv.Itoa(sh.MappedPages), list, true, true)
		sh.MappedPages++
	}
}

func (sh *submissionHandler) hasStoriesToMap() bool {
	return len(sh.Submissions)-sh.MappedSubmissions > sh.ViewableStoriesOnSinglePage
}

func createNewList(sh *submissionHandler) *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)
	setSelectedFunction(sh.Application, list, sh)

	return list
}

func addSubmissionsToList(list *cview.List, submissions []submission, sh *submissionHandler) {
	for _, submission := range submissions {
		list.AddItem(
			submission.getMainText(sh.MappedSubmissions),
			submission.getSecondaryText(),
			0,
			nil,
		)
		sh.MappedSubmissions++
	}
}

func (s submission) getMainText(i int) string {
	rank := i + 1
	return strconv.Itoa(rank) + "." + getRankIndentBlock(rank) + s.Title + s.GetDomain()
}

func (s submission) getSecondaryText() string {
	return "[::d]" + "    " + s.getPoints() + " points by " + s.Author + " " +
		s.Time + " | " + s.getComments() + " comments" + "[-:-:-]"
}

func (s submission) GetDomain() string {
	domain := s.Domain
	if domain == "" {
		return ""
	}
	return "[::d]" + " " + paren(domain) + "[-:-:-]"
}

func (s submission) getComments() string {
	return strconv.Itoa(s.CommentsCount)
}

func (s submission) getPoints() string {
	return strconv.Itoa(s.Points)
}

func paren(text string) string {
	return "(" + text + ")"
}

func getRankIndentBlock(rank int) string {
	largeIndent := "  "
	smallIndent := " "
	if rank > 9 {
		return smallIndent
	}
	return largeIndent
}
