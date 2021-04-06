package core

import (
	"gitlab.com/tslocum/cview"
)

type ScreenController struct {
	Application      *cview.Application
	Articles         *cview.List
	MainView         *MainView
	ApplicationState *ApplicationState
}

type Story struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Points        int    `json:"points"`
	Author        string `json:"user"`
	Time          int64  `json:"time"`
	CommentsCount int    `json:"comments_count"`
	URL           string `json:"url"`
	Domain        string `json:"domain"`
	Type          string `json:"type"`
}

type ApplicationState struct {
	StoriesToShow                         int
	CurrentCategory                       int
	CurrentHelpScreenCategory             int
	ScreenHeight                          int
	ScreenWidth                           int
	CurrentPage                           int
	IsReturningFromSuspension             bool
	IsOnConfigCreationConfirmationMessage bool
	IsOnAddFavoriteConfirmationMessage    bool
	IsOnDeleteFavoriteConfirmationMessage bool
	IsOnAddFavoriteByID                   bool
	State                                 int
}

type MainView struct {
	Grid           *cview.Grid
	Header         *cview.TextView
	LeftMargin     *cview.TextView
	Panels         *cview.Panels
	StatusBar      *cview.TextView
	PageCounter    *cview.TextView
	InfoScreen     *cview.TextView
	CustomFavorite *cview.InputField
}

type Config struct {
	CommentWidth        int  `mapstructure:"CLX_COMMENT_WIDTH"`
	IndentSize          int  `mapstructure:"CLX_INDENT_SIZE"`
	HighlightHeadlines  int  `mapstructure:"CLX_HIGHLIGHT_HEADLINES"`
	PreserveRightMargin bool `mapstructure:"CLX_PRESERVE_RIGHT_MARGIN"`
	RelativeNumbering   bool `mapstructure:"CLX_RELATIVE_NUMBERING"`
	HideYCJobs          bool `mapstructure:"CLX_HIDE_YC_JOBS"`
}
