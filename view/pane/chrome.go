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
	margin := strings.Repeat(" ", leftMargin)
	maxTitleWidth := max(0, screenWidth-leftMargin)
	t := syntax.ReplaceSpecialContentTags(title, enableNerdFonts)
	t = xansi.Truncate(t, maxTitleWidth, "…")

	t = syntax.HighlightYCStartupsInHeadlines(t, syntax.HeadlineInCommentSection, enableNerdFonts)
	t = syntax.HighlightYear(t, syntax.HeadlineInCommentSection)
	t = syntax.HighlightHackerNewsHeadlines(t, syntax.HeadlineInCommentSection)
	t = syntax.HighlightSpecialContent(t, syntax.HeadlineInCommentSection, enableNerdFonts)

	row := xansi.Truncate(margin+ansi.Bold+t+ansi.Reset, screenWidth, "")

	return row + "\n" + header.Underline(screenWidth)
}

func FooterSeparator(width int) string {
	s := lipgloss.NewStyle().Underline(true).Width(width)
	if header.MemorialActive() {
		s = s.Foreground(style.MemorialColor())
	}

	return s.Render(strings.Repeat(" ", width))
}
