package core

import (
	"clx/handler"
	"clx/utils/vim"

	"code.rocketnine.space/tslocum/cview"
)

type ScreenController struct {
	Application      *cview.Application
	Articles         *cview.List
	MainView         *MainView
	ApplicationState *ApplicationState
	StoryHandler     *handler.StoryHandler
	VimRegister      *vim.Register
}

type ApplicationState struct {
	StoriesToShow                         int
	CurrentCategory                       int
	ScreenHeight                          int
	ScreenWidth                           int
	CurrentPage                           int
	IsOnAddFavoriteConfirmationMessage    bool
	IsOnDeleteFavoriteConfirmationMessage bool
	IsOffline                             bool
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
	CommentWidth       int
	HighlightHeadlines bool
	HighlightComments  bool
	RelativeNumbering  bool
	HideYCJobs         bool
	AltIndentBlock     bool
	EmojiSmileys       bool
	MarkAsRead         bool
	HideIndentSymbol   bool
	OrangeHeader       bool
}

func GetConfigWithDefaults() *Config {
	return &Config{
		CommentWidth:       70,
		HighlightHeadlines: true,
		HighlightComments:  true,
		RelativeNumbering:  false,
		HideYCJobs:         true,
		AltIndentBlock:     false,
		EmojiSmileys:       true,
		MarkAsRead:         true,
		HideIndentSymbol:   false,
		OrangeHeader:       false,
	}
}
