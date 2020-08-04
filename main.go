package main

import (
	"circumflex/cmd"
	"fmt"
	"time"

	"github.com/gocolly/colly"
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
	// fmt.Println(snapshot.Stories)


	for _, s := range snapshot.Stories {
		fmt.Println(s.Title)
	}

}

type Story struct {
	Rank  string
	Title string
	URL   string
}

type topStoriesSnapshot struct {
	Stories []Story
	FoundAt time.Time
}

func CreateStory(e *colly.HTMLElement) Story {
	rank := e.ChildText(".rank")
	url := e.ChildAttr(".storylink", "href")
	title := e.ChildText(".storylink")

	return Story{Rank: rank, URL: url, Title: title}
}
