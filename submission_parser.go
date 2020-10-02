package main

import (
	"encoding/json"
	"strconv"

	"gitlab.com/tslocum/cview"
)

// SubmissionHandler stores submissions and pages
type SubmissionHandler struct {
	Submissions    []Submission
	Pages          *cview.Pages
	PagesRetrieved int
	CurrentPage    int
	StoriesListed  int
}

// Submission represents the JSON structure as
// retrieved from cheeaun's unofficial HN API
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

func fetchSubmissions(sh *SubmissionHandler) {
	sh.PagesRetrieved++
	p := strconv.Itoa(sh.PagesRetrieved)
	JSON, _ := get("http://node-hnapi.herokuapp.com/news?page=" + p)
	var submissions []Submission
	_ = json.Unmarshal(JSON, &submissions)
	sh.Submissions = append(sh.Submissions, submissions...)
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
