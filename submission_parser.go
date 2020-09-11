package main

import (
	"circumflex/client/feed"
	"strconv"

	"gitlab.com/tslocum/cview"
)

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
