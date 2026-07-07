package pane

import (
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/syntax"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// TitleHeader renders the bold, highlighted story title over an underline
// separator spanning the screen.
func TitleHeader(title string, enableNerdFonts bool, leftMargin, screenWidth int) string {
	return titleHeader(title, syntax.HeadlineInCommentSection, ansi.Bold, enableNerdFonts, leftMargin, screenWidth)
}

// LoadingTitleHeader is TitleHeader without the bold: the detail pane shows
// it while the story loads and on the error view when the load fails, so the
// title gains its full weight only once the content is in.
func LoadingTitleHeader(title string, enableNerdFonts bool, leftMargin, screenWidth int) string {
	return titleHeader(title, syntax.Unselected, "", enableNerdFonts, leftMargin, screenWidth)
}

func titleHeader(title string, highlight syntax.HighlightType, baseStyle string, enableNerdFonts bool, leftMargin, screenWidth int) string {
	margin := strings.Repeat(" ", leftMargin)
	maxTitleWidth := max(0, screenWidth-leftMargin)
	t := syntax.ReplaceSpecialContentTags(title, enableNerdFonts)
	t = xansi.Truncate(t, maxTitleWidth, "…")

	t = syntax.HighlightYCStartupsInHeadlines(t, highlight, enableNerdFonts)
	t = syntax.HighlightYear(t, highlight)
	t = syntax.HighlightHackerNewsHeadlines(t, highlight)
	t = syntax.HighlightSpecialContent(t, highlight, enableNerdFonts)

	row := xansi.Truncate(margin+baseStyle+t+ansi.Reset, screenWidth, "")

	return row + "\n" + header.Underline(screenWidth)
}

func FooterSeparator(width int) string {
	s := lipgloss.NewStyle().Underline(true).Width(width)
	if header.MemorialActive() {
		s = s.Foreground(style.MemorialColor())
	}

	return s.Render(strings.Repeat(" ", width))
}
