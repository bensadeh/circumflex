package settings

import "github.com/bensadeh/circumflex/theme"

const (
	defaultCommentWidth   = 70
	defaultArticleWidth   = 80
	defaultPageMultiplier = 3
	defaultIndent         = 1
	minIndent             = 1
	minPageMultiplier     = 1
	maxPageMultiplier     = 5
)

type Config struct {
	CommentWidth               int
	ArticleWidth               int
	Indent                     int
	DoNotMarkSubmissionsAsRead bool
	DebugMode                  bool
	DebugFallible              bool
	EnableNerdFonts            bool
	PageMultiplier             int
	Theme                      *theme.Theme
}

func Default() *Config {
	return &Config{
		CommentWidth:   defaultCommentWidth,
		ArticleWidth:   defaultArticleWidth,
		Indent:         defaultIndent,
		PageMultiplier: defaultPageMultiplier,
	}
}

func ClampPageMultiplier(n int) int {
	return max(minPageMultiplier, min(n, maxPageMultiplier))
}

func ClampIndent(n int) int {
	return max(minIndent, n)
}
