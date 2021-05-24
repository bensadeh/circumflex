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
	AltIndentBlock      bool `mapstructure:"CLX_ALT_INDENT_BLOCK"`
}
