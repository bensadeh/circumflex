package settings

import "clx/theme"

const defaultCommentWidth = 70

type Config struct {
	CommentWidth                int
	DisableHeadlineHighlighting bool
	DisableCommentHighlighting  bool
	DisableEmojis               bool
	DoNotMarkSubmissionsAsRead  bool
	HideIndentSymbol            bool
	IndentationSymbol           string
	DebugMode                   bool
	DebugFallible               bool
	EnableNerdFonts             bool
	LesskeyPath                 string
	AutoExpandComments          bool
	NoLessVerify                bool
	Theme                       *theme.Theme
}

func Default() *Config {
	return &Config{
		CommentWidth:      defaultCommentWidth,
		IndentationSymbol: " ▎",
	}
}
