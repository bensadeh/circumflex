package main

import (
	"strconv"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
	"gitlab.com/tslocum/cview"
)

// Submission represents the JSON structure as
// retreived from cheeaun's unoffical HN API
type Submission struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Points        int    `json:"points"`
	Author        string `json:"user"`
	Time          string `json:"time_ago"`
	CommentsCount int    `json:"comments_count"`
	URL           string `json:"url"`
	Domain        string `json:"domain"`
	Type          string `json:"type"`
}

func addListItems(list *cview.List, app *cview.Application, sub []Submission) {
	y, _ := terminal.Height()
	storiesToFetch := int(y / 2)

	for i := 0; i < storiesToFetch; i++ {
		rank := i + 1
		indentedRank := strconv.Itoa(rank) + "." + getRankIndentBlock(rank)
		primary := indentedRank + sub[i].Title + getDomain(sub[i].Domain)
		points := strconv.Itoa(sub[i].Points)
		comments := strconv.Itoa(sub[i].CommentsCount)
		secondary := "    " + points + " points by " + sub[i].Author + " " + sub[i].Time + " | " + comments + " comments"
		list.AddItem(primary, secondary, 0, nil)
	}

}

func getDomain(domain string) string {
	if domain == "" {
		return ""
	}
	return "[::d]" + " " + paren(domain) + "[-:-:-]"
}

func getRankIndentBlock(rank int) string {
	largeIndent := "  "
	smallIndent := " "
	if rank > 9 {
		return smallIndent
	}
	return largeIndent
}
