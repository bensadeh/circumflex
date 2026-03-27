package settings

import "clx/theme"

const (
	defaultCommentWidth   = 70
	defaultPageMultiplier = 3
	minPageMultiplier     = 1
	maxPageMultiplier     = 5
)

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
	NativeCommentView           bool
	PageMultiplier              int
	Theme                       *theme.Theme
}

func Default() *Config {
	return &Config{
		CommentWidth:      defaultCommentWidth,
		PageMultiplier:    defaultPageMultiplier,
		IndentationSymbol: " ▎",
	}
}

func ClampPageMultiplier(n int) int {
	return max(minPageMultiplier, min(n, maxPageMultiplier))
}
