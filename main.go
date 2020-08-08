package main

import (
	"circumflex/cmd"
	"time"

	"github.com/gdamore/tcell"
	"github.com/gocolly/colly"
	"gitlab.com/tslocum/cview"
)

func main() {
	cmd.Execute()

	c := colly.NewCollector()
	snapshot := topStoriesSnapshot{FoundAt: time.Now()}

	c.OnHTML(".athing", func(e *colly.HTMLElement) {
		story := CreateStory(e)
		snapshot.Stories = append(snapshot.Stories, story)
	})

	c.Visit("https://news.ycombinator.com")

	// for _, s := range snapshot.Stories {
	// 	fmt.Println(s.Username)
	// }

	app := cview.NewApplication()
	list := cview.NewList()

	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorGray)

	reset := func() {
		list.Clear()
		for _, s := range snapshot.Stories {
			list.AddItem(s.Title, s.Username, 0, nil)
		}
	}

	reset()
	if err := app.SetRoot(list, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}

type Story struct {
	Rank     string
	Title    string
	URL      string
	Username string
}

type topStoriesSnapshot struct {
	Stories []Story
	FoundAt time.Time
}

func CreateStory(e *colly.HTMLElement) Story {
	rank := e.ChildText(".rank")
	url := e.ChildAttr(".storylink", "href")
	title := e.ChildText(".storylink")
	metaDataRow := e.DOM.Next()
	username := metaDataRow.Find(".hnuser").Text()

	return Story{Rank: rank, URL: url, Title: title, Username: username}
}
