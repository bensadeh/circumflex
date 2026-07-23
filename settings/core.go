package settings

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/graphics"
	"github.com/bensadeh/circumflex/theme"
)

const (
	defaultCommentWidth   = 70
	defaultArticleWidth   = 80
	defaultPageMultiplier = 3
	defaultIndent         = 1
	minIndent             = 1
	minWidth              = 1
	minPageMultiplier     = 1
	maxPageMultiplier     = 5

	// DefaultWideViewMinWidth is the terminal width at which the wide
	// (split-pane) layout kicks in unless configured otherwise.
	DefaultWideViewMinWidth = 180
)

type Config struct {
	CommentWidth               int
	ArticleWidth               int
	Indent                     int
	DoNotMarkSubmissionsAsRead bool
	DebugMode                  bool
	DebugFallible              bool
	EnableNerdFonts            bool
	ShowImagesOnOpen           bool
	Graphics                   graphics.Mode
	PageMultiplier             int
	WideViewMinWidth           int
	Categories                 string
	Theme                      *theme.Theme
}

func Default() *Config {
	return &Config{
		CommentWidth:     defaultCommentWidth,
		ArticleWidth:     defaultArticleWidth,
		Indent:           defaultIndent,
		PageMultiplier:   defaultPageMultiplier,
		WideViewMinWidth: DefaultWideViewMinWidth,
		Categories:       categories.Default,
	}
}

// ParseWideView converts a wide-view setting into the minimum terminal
// width for the split-pane layout: "always" enables it at any width,
// "never" disables it, and a number enables it from that many columns on.
// The error carries no flag or config-key context; callers add their own.
func ParseWideView(value string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "never":
		return math.MaxInt, nil
	case "always":
		return 0, nil
	}

	width, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || width < 1 {
		return 0, fmt.Errorf(`must be "always", "never" or a column count, got %q`, value)
	}

	return width, nil
}

// ParseGraphics converts a graphics setting into a detection mode: "auto"
// lets the terminal's answer decide, "always" and "never" overrule it. The
// error carries no flag or config-key context; callers add their own.
func ParseGraphics(value string) (graphics.Mode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "auto":
		return graphics.ModeAuto, nil
	case "always":
		return graphics.ModeAlways, nil
	case "never":
		return graphics.ModeNever, nil
	}

	return graphics.ModeAuto, fmt.Errorf(`must be "auto", "always" or "never", got %q`, value)
}

func ClampPageMultiplier(n int) int {
	return max(minPageMultiplier, min(n, maxPageMultiplier))
}

func ClampIndent(n int) int {
	return max(minIndent, n)
}

// ClampWidth floors the comment- and article-width settings. Only a floor:
// the terminal's own width caps them from above at render time.
func ClampWidth(n int) int {
	return max(minWidth, n)
}
