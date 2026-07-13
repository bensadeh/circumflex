package pane

import (
	"strings"

	"github.com/bensadeh/circumflex/header"
	"github.com/bensadeh/circumflex/headline"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// TitleHeader renders the bold, highlighted story title over an underline
// separator spanning the screen.
func TitleHeader(title string, enableNerdFonts bool, leftMargin, screenWidth int) string {
	return titleHeader(title, headline.HeadlineInCommentSection, enableNerdFonts, leftMargin, screenWidth)
}

// LoadingTitleHeader is TitleHeader without the bold: the detail pane shows
// it while the story loads and on the error view when the load fails, so the
// title gains its full weight only once the content is in.
func LoadingTitleHeader(title string, enableNerdFonts bool, leftMargin, screenWidth int) string {
	return titleHeader(title, headline.Unselected, enableNerdFonts, leftMargin, screenWidth)
}

func titleHeader(title string, highlight headline.HighlightType, enableNerdFonts bool, leftMargin, screenWidth int) string {
	margin := strings.Repeat(" ", leftMargin)
	maxTitleWidth := max(0, screenWidth-leftMargin)

	// Render the full title first, truncate the styled output after: the
	// escape-aware cut cannot strand a half-eaten token pattern.
	t := headline.Render(title, highlight, enableNerdFonts)
	t = xansi.Truncate(t, maxTitleWidth, "…")

	row := xansi.Truncate(margin+t, screenWidth, "")

	return row + "\n" + header.Underline(screenWidth)
}

// TitleHeaderWithBadge is TitleHeader with a pre-styled badge ending at
// rightEdge — the content column's right edge, not the screen's. The title
// truncates early enough that the two never collide. A pane too narrow to
// keep any title next to the badge drops the badge instead.
func TitleHeaderWithBadge(title, badge string, enableNerdFonts bool, leftMargin, rightEdge, screenWidth int) string {
	const badgeGap = 2

	badgeWidth := xansi.StringWidth(badge)

	maxTitleWidth := rightEdge - leftMargin - badgeWidth - badgeGap
	if maxTitleWidth < 1 {
		return TitleHeader(title, enableNerdFonts, leftMargin, screenWidth)
	}

	t := headline.Render(title, headline.HeadlineInCommentSection, enableNerdFonts)
	t = xansi.Truncate(t, maxTitleWidth, "…")

	pad := rightEdge - leftMargin - xansi.StringWidth(t) - badgeWidth

	row := strings.Repeat(" ", leftMargin) + t + strings.Repeat(" ", pad) + badge
	row = xansi.Truncate(row, screenWidth, "")

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
