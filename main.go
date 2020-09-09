package main

import (
	"circumflex/client"
	"circumflex/client/feed"
	"circumflex/cmd"
	"fmt"

	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

type Foo struct {
	Bar string
}

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

			for index, _ := range *pp {
				if index == i {
					foo1 := new(Foo) // or &Foo{}
					JSON, _ := getJSON("http://node-hnapi.herokuapp.com/item/3338485", foo1)
					println(JSON)

					// _ := scrapeComments(s.ID)
					outputStringToLess(string(JSON))
				}
			}
		})
	})

	addListItems(list, pp, app)
	if err := app.SetRoot(list, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}

}

func addListItems(list *cview.List, pp *[]feed.Item, app *cview.Application) {
	list.Clear()
	for i, s := range *pp {
		rank := i + 1
		indentedRank := strconv.Itoa(rank) + "." + getRankIndentBlock(rank)
		points := strconv.Itoa(s.Points)
		comments := strconv.Itoa(s.Comments)
		secondary := "    " + points + " points by " + s.Author + " " + s.Age + " | " + comments + " comments"
		list.AddItem(indentedRank+s.Title, secondary, 0, nil)
	}
}

func getRankIndentBlock(rank int) string {
	largeIndent := "  "
	smallIndent := " "
	if rank > 9 {
		return smallIndent
	}
	return largeIndent
}

func outputStringToLess(output string) {
	cmd := exec.Command("less", "-R")
	cmd.Stdin = strings.NewReader(output)
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
