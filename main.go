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
	"gitlab.com/tslocum/cview"
)

func main() {
	cmd.Execute()

	JSON, _ := get("http://node-hnapi.herokuapp.com/news?page=1")

	var jSubmission []Submission
	json.Unmarshal(JSON, &jSubmission)

	app := cview.NewApplication()
	list := cview.NewList()

	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorGray)
	list.ShowSecondaryText(true)
	list.SetSelectedFunc(func(i int, a string, b string, c rune) {
		app.Suspend(func() {
			// Clear screen to avoid seeing text between
			// viewing submissions and comments
			clearScreen()

			for index := range jSubmission {
				if index == i {
					id := strconv.Itoa(jSubmission[i].ID)
					JSON, _ := get("http://node-hnapi.herokuapp.com/item/" + id)
					var jComments = new(Comments)
					json.Unmarshal(JSON, jComments)
					originalPoster := jSubmission[i].Author
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

	addListItems(list, app, jSubmission)
	if err := app.SetRoot(list, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}

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
