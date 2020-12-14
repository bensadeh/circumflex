package types

import "gitlab.com/tslocum/cview"

const (
	NoCategory  = 0
	New         = 1
	Ask         = 2
	Show        = 3
	NewsPanel   = "0"
	NewestPanel = "1"
	AskPanel    = "2"
	ShowPanel   = "3"
)

type ScreenController struct {
	Application      *cview.Application
	MainView         *MainView
	Submissions      []*Submissions
	ApplicationState *ApplicationState
}

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

type Submissions struct {
	MappedSubmissions  int
	MappedPages        int
	StoriesListed      int
	PageToFetchFromAPI int
	MaxPages           int
	SubmissionEntries  []*Submission
}

type ApplicationState struct {
	ViewableStoriesOnSinglePage int
	CurrentCategory             int
	ScreenHeight                int
	ScreenWidth                 int
	CurrentPage                 int
	IsOffline                   bool
	IsReturningFromSuspension   bool
	IsOnHelpScreen              bool
}

type MainView struct {
	Panels      *cview.Panels
	Grid        *cview.Grid
	Footer      *cview.TextView
	Header      *cview.TextView
	LeftMargin  *cview.TextView
	RightMargin *cview.TextView
}
