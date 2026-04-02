package settings

import "github.com/bensadeh/circumflex/theme"

const (
	defaultCommentWidth   = 70
	defaultArticleWidth   = 80
	defaultPageMultiplier = 3
	minPageMultiplier     = 1
	maxPageMultiplier     = 5
)

type Config struct {
	CommentWidth               int
	ArticleWidth               int
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
		PageMultiplier: defaultPageMultiplier,
	}
}

func ClampPageMultiplier(n int) int {
	return max(minPageMultiplier, min(n, maxPageMultiplier))
}
