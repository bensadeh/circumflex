package structs

import (
	"gitlab.com/tslocum/cview"
)

type ScreenController struct {
	Application      *cview.Application
	Articles         *cview.List
	Settings         *Settings
	SettingsHeader   *cview.TextView
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
	Entries            []*Submission
}

type ApplicationState struct {
	SubmissionsToShow         int
	SubmissionsCategory       int
	SettingsCategory          int
	HelpScreenCategory        int
	ScreenHeight              int
	ScreenWidth               int
	CurrentPage               int
	IsOffline                 bool
	IsReturningFromSuspension bool
	IsOnHelpScreen            bool
}

type MainView struct {
	Grid   *cview.Grid
	Header *cview.TextView
	//Put Settings Header here
	LeftMargin  *cview.TextView
	Panels      *cview.Panels
	StatusBar   *cview.TextView
	PageCounter *cview.TextView
}

type Settings struct {
	List                  *cview.List
	CurrentSettingsPage   int
	NumberOfSettingsPages int
}
