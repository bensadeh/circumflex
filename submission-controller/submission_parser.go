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

// SubmissionHandler stores submissions and pages
type SubmissionHandler struct {
	Submissions                 []Submission
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

func NewSubmissionHandler() *SubmissionHandler {
	sh := new(SubmissionHandler)
	sh.Application = cview.NewApplication()
	sh.setShortcuts()
	sh.Pages = cview.NewPages()
	sh.MaxPages = 2
	sh.ScreenHeight = getTerminalHeight()
	sh.ViewableStoriesOnSinglePage = min(sh.ScreenHeight/2, maximumStoriesToDisplay)
	sh.FetchSubmissions()

	sh.Pages.SwitchToPage("0")
	return sh
}

func (sh *SubmissionHandler) setShortcuts() {
	app := sh.Application
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			sh.NextPage()
		} else if event.Key() == tcell.KeyCtrlP {
			sh.PreviousPage()
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

func (sh *SubmissionHandler) NextPage() {
	nextPage := sh.CurrentPage + 1

	if nextPage > sh.MaxPages {
		return
	}

	if nextPage < sh.MappedPages {
		sh.CurrentPage++
		sh.Pages.SwitchToPage(strconv.Itoa(sh.CurrentPage))
	} else {
		sh.FetchSubmissions()
		sh.CurrentPage++
		sh.Pages.SwitchToPage(strconv.Itoa(sh.CurrentPage))
	}
}

func (sh *SubmissionHandler) PreviousPage() {
	previousPage := sh.CurrentPage - 1

	if previousPage < 0 {
		return
	}

	sh.CurrentPage--
	sh.Pages.SwitchToPage(strconv.Itoa(sh.CurrentPage))
}

func (sh *SubmissionHandler) GetStoriesToDisplay() int {
	return sh.ViewableStoriesOnSinglePage
}

func setSelectedFunction(app *cview.Application, list *cview.List, sh *SubmissionHandler) {
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

func (sh *SubmissionHandler) GetSubmissionInfo(i int) (string, string) {
	submission := sh.GetSubmission(i)

	rank := i + 1
	indentedRank := strconv.Itoa(rank) + "." + GetRankIndentBlock(rank)

	primary := indentedRank + submission.Title + submission.GetDomain()

	secondary := "[::d]" + "    " + submission.GetPoints() + " points by " + submission.Author + " " + submission.Time + " | " + submission.GetComments() + " comments" + "[-:-:-]"

	return primary, secondary
}

func (sh *SubmissionHandler) GetSubmission(i int) Submission {
	return sh.Submissions[i]
}

// Submission represents the JSON structure as
// retrieved from cheeaun's unofficial HN API
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

func (sh *SubmissionHandler) FetchSubmissions() {
	sh.PageToFetchFromAPI++
	p := strconv.Itoa(sh.PageToFetchFromAPI)
	JSON, _ := http.Get("http://node-hnapi.herokuapp.com/news?page=" + p)
	var submissions []Submission
	_ = json.Unmarshal(JSON, &submissions)
	sh.Submissions = append(sh.Submissions, submissions...)
	sh.mapSubmissions()
}

func (sh *SubmissionHandler) mapSubmissions() {
	for sh.hasStoriesToMap() {
		sub := sh.Submissions[sh.MappedSubmissions : sh.MappedSubmissions+sh.ViewableStoriesOnSinglePage]
		list := createNewList(sh)
		addSubmissionsToList(list, sub, sh)

		sh.Pages.AddPage(strconv.Itoa(sh.MappedPages), list, true, true)
		sh.MappedPages++
	}
}

func (sh *SubmissionHandler) hasStoriesToMap() bool {
	return len(sh.Submissions)-sh.MappedSubmissions > sh.ViewableStoriesOnSinglePage
}

func createNewList(sh *SubmissionHandler) *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)
	setSelectedFunction(sh.Application, list, sh)

	return list
}

func addSubmissionsToList(list *cview.List, submissions []Submission, sh *SubmissionHandler) {

	for _, submission := range submissions {
		list.AddItem(submission.getMainText(sh.MappedSubmissions), submission.getSecondaryText(), 0, nil)
		sh.MappedSubmissions++
	}
}

func (s Submission) getMainText(i int) string {
	rank := i + 1
	indentedRank := strconv.Itoa(rank) + "." + GetRankIndentBlock(rank)

	return indentedRank + s.Title + s.GetDomain()
}

func (s Submission) getSecondaryText() string {
	return "[::d]" + "    " + s.GetPoints() + " points by " + s.Author + " " +
		s.Time + " | " + s.GetComments() + " comments" + "[-:-:-]"
}

func (s Submission) GetDomain() string {
	domain := s.Domain
	if domain == "" {
		return ""
	}
	return "[::d]" + " " + paren(domain) + "[-:-:-]"
}

func (s Submission) GetComments() string {
	return strconv.Itoa(s.CommentsCount)
}

func (s Submission) GetPoints() string {
	return strconv.Itoa(s.Points)
}

func paren(text string) string {
	return "(" + text + ")"
}

func GetRankIndentBlock(rank int) string {
	largeIndent := "  "
	smallIndent := " "
	if rank > 9 {
		return smallIndent
	}
	return largeIndent
}
