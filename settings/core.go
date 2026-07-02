package settings

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/theme"
)

const (
	defaultCommentWidth   = 70
	defaultArticleWidth   = 80
	defaultPageMultiplier = 3
	defaultIndent         = 1
	minIndent             = 1
	minPageMultiplier     = 1
	maxPageMultiplier     = 5

	// DefaultWideViewMinWidth is the terminal width at which the wide
	// (split-pane) layout kicks in unless configured otherwise.
	DefaultWideViewMinWidth = 240
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
	WideViewMinWidth           int
	Theme                      *theme.Theme
}

func Default() *Config {
	return &Config{
		CommentWidth:     defaultCommentWidth,
		ArticleWidth:     defaultArticleWidth,
		Indent:           defaultIndent,
		PageMultiplier:   defaultPageMultiplier,
		WideViewMinWidth: DefaultWideViewMinWidth,
	}
}

// ParseWideView converts the --wide-view flag value into the minimum
// terminal width for the split-pane layout: "always" enables it at any
// width, "never" disables it, and a number enables it from that many
// columns on.
func ParseWideView(value string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "never":
		return math.MaxInt, nil
	case "always":
		return 0, nil
	}

	width, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || width < 1 {
		return 0, fmt.Errorf(`--wide-view must be "always", "never" or a column count, got %q`, value)
	}

	return width, nil
}

func ClampPageMultiplier(n int) int {
	return max(minPageMultiplier, min(n, maxPageMultiplier))
}

func ClampIndent(n int) int {
	return max(minIndent, n)
}
