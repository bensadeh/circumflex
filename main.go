package main

import (
	"circumflex/client"
	"circumflex/cmd"
	"encoding/json"
	"fmt"

	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func main() {
	cmd.Execute()
	y, _ := terminal.Height()
	storiesToFetch := int(y / 2)

	client := client.NewHNClient()
	pp, err := client.GetTopStories(storiesToFetch)
	if err != nil {
		fmt.Println(err)
		return
	}

	app := cview.NewApplication()
	list := cview.NewList()

	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorGray)
	list.ShowSecondaryText(true)
	list.SetSelectedFunc(func(i int, a string, b string, c rune) {
		app.Suspend(func() {
			//Clear screen to avoid seeing the terminal before
			//this program was started
			c := exec.Command("clear")
			c.Stdout = os.Stdout
			c.Run()

			for index, s := range *pp {
				if index == i {
					JSON, _ := get("http://node-hnapi.herokuapp.com/item/" + s.ID)
					var jComments = new(Comments)
					json.Unmarshal(JSON, jComments)
					originalPoster := s.Author
					commentTree := ""
					appendCommentsHeader(*jComments, &commentTree)
					for _, s := range jComments.Replies {
						commentTree = prettyPrintComments(*s, &commentTree, 0, originalPoster)
					}

					outputStringToLess(commentTree)
				}
			}
		})
	})

	addListItems(list, pp, app)
	if err := app.SetRoot(list, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}

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
