package main

import (
	"clx/cmd"
	"clx/comment-parser"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"

	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
	"gitlab.com/tslocum/cview"
)

func main() {
	cmd.Execute()
	clearScreen()
	submissionHandler := new(SubmissionHandler)

	app := cview.NewApplication()
	pages := cview.NewPages()
	submissionHandler.Pages = pages

	initNewPage(app, submissionHandler)
	submissionHandler.Pages.SwitchToPage("0")

	setShortcuts(app, submissionHandler)

	if err := app.SetRoot(pages, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}

}

func setShortcuts(app *cview.Application, submissionHandler *SubmissionHandler) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			nextSlide(app, submissionHandler)
		} else if event.Key() == tcell.KeyCtrlP {
			submissionHandler.Pages.SwitchToPage("0")
		} else if event.Rune() == 'q' {
			app.Stop()
		}
		return event
	})
}

func nextSlide(app *cview.Application, sh *SubmissionHandler) {
	sh.CurrentPage++
	initNewPage(app, sh)
	sh.Pages.SwitchToPage("1")
	// currentPage := strconv.Itoa(sh.CurrentPage)
	// sh.Pages.SwitchToPage(currentPage)

	// app.SetRoot(pageToView, true)
	// panic(sh.CurrentPage)
}

func initNewPage(app *cview.Application, sh *SubmissionHandler) {
	// y, _ := terminal.Height()
	// storiesToView := int(y/2) * (sh.CurrentPage + 1)
	// availableSubmissions := len(sh.Submissions)

	fetchSubmissions(sh)

	list := createNewList(app, sh)
	// sh.Pages = append(sh.Pages, newPage)

	currentPage := strconv.Itoa(sh.CurrentPage)
	sh.Pages.AddPage(currentPage, list, true, true)
}

func createNewList(app *cview.Application, sh *SubmissionHandler) *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)
	setSelectedFunction(app, list, sh)

	addListItems(list, sh)

	return list
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
					JSON, _ := get("http://node-hnapi.herokuapp.com/item/" + id)
					var jComments = new(comment_parser.Comments)
					_ = json.Unmarshal(JSON, jComments)

					commentTree := comment_parser.PrintCommentTree(*jComments, 3, 70)
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

func addListItems(list *cview.List, sh *SubmissionHandler) {
	y, _ := terminal.Height()
	storiesToShow := int(y/2) * (sh.CurrentPage + 1)

	for i := sh.StoriesListed; i < storiesToShow; i++ {
		sh.StoriesListed++
		primary, secondary := getSubmissionInfo(i, sh.Submissions[i])
		list.AddItem(primary, secondary, 0, nil)
	}
}

func getSubmissionInfo(i int, submission Submission) (string, string) {
	rank := i + 1
	indentedRank := strconv.Itoa(rank) + "." + getRankIndentBlock(rank)
	primary := indentedRank + submission.Title + getDomain(submission.Domain)
	points := strconv.Itoa(submission.Points)
	comments := strconv.Itoa(submission.CommentsCount)
	secondary := "[::d]" + "    " + points + " points by " + submission.Author + " " + submission.Time + " | " + comments + " comments" + "[-:-:-]"
	return primary, secondary
}

func clearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
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
