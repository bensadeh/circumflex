package settings

import "clx/theme"

const (
	defaultCommentWidth   = 70
	defaultArticleWidth   = 80
	defaultPageMultiplier = 3
	minPageMultiplier     = 1
	maxPageMultiplier     = 5
)

type Config struct {
	CommentWidth                int
	ArticleWidth                int
	DisableHeadlineHighlighting bool
	DisableCommentHighlighting  bool
	DisableEmojis               bool
	DoNotMarkSubmissionsAsRead  bool
	IndentationSymbol           string
	DebugMode                   bool
	DebugFallible               bool
	EnableNerdFonts             bool
	PageMultiplier              int
	Theme                       *theme.Theme
}

func Default() *Config {
	return &Config{
		CommentWidth:      defaultCommentWidth,
		ArticleWidth:      defaultArticleWidth,
		PageMultiplier:    defaultPageMultiplier,
		IndentationSymbol: " ▎",
	}
}

func ClampPageMultiplier(n int) int {
	return max(minPageMultiplier, min(n, maxPageMultiplier))
}
