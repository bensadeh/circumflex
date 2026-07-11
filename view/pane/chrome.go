package pane

import (
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/headline"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// TitleHeader renders the bold, highlighted story title over an underline
// separator spanning the screen.
func TitleHeader(title string, enableNerdFonts bool, leftMargin, screenWidth int) string {
	return titleHeader(title, headline.HeadlineInCommentSection, ansi.Bold, enableNerdFonts, leftMargin, screenWidth)
}

// LoadingTitleHeader is TitleHeader without the bold: the detail pane shows
// it while the story loads and on the error view when the load fails, so the
// title gains its full weight only once the content is in.
func LoadingTitleHeader(title string, enableNerdFonts bool, leftMargin, screenWidth int) string {
	return titleHeader(title, headline.Unselected, "", enableNerdFonts, leftMargin, screenWidth)
}

func titleHeader(title string, highlight headline.HighlightType, baseStyle string, enableNerdFonts bool, leftMargin, screenWidth int) string {
	margin := strings.Repeat(" ", leftMargin)
	maxTitleWidth := max(0, screenWidth-leftMargin)
	t := headline.ReplaceSpecialContentTags(title, enableNerdFonts)
	t = xansi.Truncate(t, maxTitleWidth, "…")

	t = headline.HighlightYCStartupsInHeadlines(t, highlight, enableNerdFonts)
	t = headline.HighlightYear(t, highlight)
	t = headline.HighlightHackerNewsHeadlines(t, highlight)
	t = headline.HighlightSpecialContent(t, highlight, enableNerdFonts)

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

// FooterSections spreads footer labels across width: the first sits flush
// left, the last ends flush right, and the space left over is shared
// equally by the gaps between them. Labels that together overrun the width
// keep a single space between them instead; the caller truncates.
func FooterSections(width int, sections ...string) string {
	gaps := max(1, len(sections)-1)

	slack := width
	for _, s := range sections {
		slack -= xansi.StringWidth(s)
	}

	var b strings.Builder

	cur, prefix := 0, 0

	for i, s := range sections {
		if s == "" {
			continue
		}

		pad := prefix + slack*i/gaps - cur
		if b.Len() > 0 {
			pad = max(pad, 1)
		} else {
			pad = max(pad, 0)
		}

		b.WriteString(strings.Repeat(" ", pad) + s)

		w := xansi.StringWidth(s)
		cur += pad + w
		prefix += w
	}

	return b.String()
}
