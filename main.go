package main

import (
	"clx/cmd"
	"encoding/json"
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
	initNewPage(app, submissionHandler)

	// Shortcuts to navigate the slides.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			nextSlide(app, submissionHandler)
		} else if event.Key() == tcell.KeyCtrlP {
			// previousSlide()
		}
		return event
	})

	firstPage := submissionHandler.Pages[0]
	if err := app.SetRoot(firstPage, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}

}

func nextSlide(app *cview.Application, sh *SubmissionHandler) {
	initNewPage(app, sh)
	sh.CurrentPage++
	pageToView := sh.Pages[sh.CurrentPage]

	app.SetRoot(pageToView, true)
	panic(sh.CurrentPage)
}

func initNewPage(app *cview.Application, sh *SubmissionHandler) {
	y, _ := terminal.Height()
	storiesToView := int(y/2) * (sh.CurrentPage + 1)
	availableSubmissions := len(sh.Submissions)

	if storiesToView > availableSubmissions {
		fetchSubmissions(sh)
	}

	newPage := createNewList(app, sh)
	sh.Pages = append(sh.Pages, newPage)
}

func createNewList(app *cview.Application, sh *SubmissionHandler) *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorGray)
	list.ShowSecondaryText(true)
	setSelectedFunction(app, list, sh)

	addListItems(list, app, sh)

	return list
}

func setSelectedFunction(app *cview.Application, list *cview.List, sh *SubmissionHandler) {
	list.SetSelectedFunc(func(i int, a string, b string, c rune) {
		app.Suspend(func() {
			for index := range sh.Submissions {
				if index == i {
					id := strconv.Itoa(sh.Submissions[i].ID)
					JSON, _ := get("http://node-hnapi.herokuapp.com/item/" + id)
					var jComments = new(Comments)
					json.Unmarshal(JSON, jComments)
					originalPoster := sh.Submissions[i].Author
					commentTree := ""
					appendCommentsHeader(*jComments, &commentTree)
					for _, s := range jComments.Replies {
						commentTree = prettyPrintComments(*s, &commentTree, 0, 5, 70, originalPoster)
					}

					outputStringToLess(commentTree)
				}
			}
		})
	})
}

func addListItems(list *cview.List, app *cview.Application, sh *SubmissionHandler) {
	y, _ := terminal.Height()
	storiesToShow := int(y/2) * (sh.CurrentPage + 1)
	startCounter := storiesToShow*(sh.CurrentPage+1) - storiesToShow

	for i := startCounter; i < storiesToShow; i++ {
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
	secondary := "    " + points + " points by " + submission.Author + " " + submission.Time + " | " + comments + " comments"
	return primary, secondary
}

func clearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	c.Run()
}

func outputStringToLess(output string) {
	cmd := exec.Command("less", "-r")
	cmd.Stdin = strings.NewReader(output)
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
