package main

import (
	"circumflex/client"
	"circumflex/cmd"
	"fmt"

	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
)

func main() {
	cmd.Execute()

	client := client.NewHNClient()
	pp, err := client.GetTopStories(30)
	if err != nil {
		fmt.Println(err)
		return
	}

	// for _, v := range *pp {
	// 	fmt.Println(v.Title)
	// }

	app := cview.NewApplication()
	list := cview.NewList()

	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorGray)
	list.ShowSecondaryText(false)

	reset := func() {
		list.Clear()
		for _, s := range *pp {
			list.AddItem(s.Title, s.Author, 0, nil)
		}
	}

	reset()
	if err := app.SetRoot(list, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
