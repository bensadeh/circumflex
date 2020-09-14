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
	submissionHandler.Submissions = fetchSubmissions(1)

	app := cview.NewApplication()
	list := getNewList(app, submissionHandler.Submissions)
	submissionHandler.Pages = append(submissionHandler.Pages, list)
	secondList := cview.NewList()

	// Shortcuts to navigate the slides.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlN {
			// nextSlide()
		} else if event.Key() == tcell.KeyCtrlP {
			// previousSlide()
		}
		return event
	})

	addListItems(submissionHandler.Pages[0], app, submissionHandler.Submissions, secondList)
	if err := app.SetRoot(submissionHandler.Pages[0], true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}

}

func getNewList(app *cview.Application, submissions []Submission) *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorGray)
	list.ShowSecondaryText(true)
	list.SetSelectedFunc(func(i int, a string, b string, c rune) {
		app.Suspend(func() {
			for index := range submissions {
				if index == 16 {
					return
				}
				if index == i {
					id := strconv.Itoa(submissions[i].ID)
					JSON, _ := get("http://node-hnapi.herokuapp.com/item/" + id)
					var jComments = new(Comments)
					json.Unmarshal(JSON, jComments)
					originalPoster := submissions[i].Author
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

	return list
}

func addListItems(list *cview.List, app *cview.Application, sub []Submission, secondList *cview.List) {
	y, _ := terminal.Height()
	storiesToFetch := int(y/2) - 1

	for i := 0; i < storiesToFetch; i++ {
		primary, secondary := getSubmissionInfo(i, sub[i])
		list.AddItem(primary, secondary, 0, nil)
	}

	list.AddItem("More", "", 0, func() {
		for i := storiesToFetch; i < 30; i++ {
			primary, secondary := getSubmissionInfo(i, sub[i])
			secondList.AddItem(primary, secondary, 0, nil)
		}
		app.SetRoot(secondList, true)
	})

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
