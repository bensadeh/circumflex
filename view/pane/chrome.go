package pane

import (
	"image/color"
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

// DepthBadge marks a page reached by following links in place: one faint
// chevron per link behind it, each one a quit-key step back. The floor keeps
// a linked page with no chain at all from rendering an empty badge.
func DepthBadge(depth int) string {
	return style.Faint(strings.Repeat("›", max(1, depth)))
}

func FooterSeparator(width int) string {
	return footerRule().Width(width).Render(strings.Repeat(" ", width))
}

// FooterSeparatorWithLabel is FooterSeparator with label written into the
// rule, ending at rightEdge. The label's glyphs render dim while the
// underline through them keeps the rule's full strength: the underline color
// is pinned to the rule's own (SGR 58) and the text dims through an
// explicitly blended foreground — SGR faint cannot deliver this, terminals
// dim the underline decoration along with the glyphs no matter the underline
// color. termFG/termBG are the terminal's reported colors; without them the
// blend has no inputs and the label falls back to faint, dimming its stretch
// of the rule with it.
func FooterSeparatorWithLabel(width, rightEdge int, label string, termFG, termBG color.Color) string {
	rightEdge = min(rightEdge, width)
	label = xansi.Truncate(label, max(0, rightEdge), "…")

	ruleFG := termFG
	if header.MemorialActive() {
		ruleFG = style.MemorialColor()
	}

	labelStyle := footerRule()

	switch {
	case ruleFG != nil && termBG != nil:
		labelStyle = labelStyle.Foreground(style.Dimmed(ruleFG, termBG)).UnderlineColor(ruleFG)
	case ruleFG != nil:
		labelStyle = labelStyle.Faint(true).UnderlineColor(ruleFG)
	default:
		labelStyle = labelStyle.Faint(true)
	}

	lead := rightEdge - xansi.StringWidth(label)

	return footerRule().Render(strings.Repeat(" ", lead)) +
		labelStyle.Render(label) +
		footerRule().Render(strings.Repeat(" ", width-rightEdge))
}

func footerRule() lipgloss.Style {
	s := lipgloss.NewStyle().Underline(true)
	if header.MemorialActive() {
		s = s.Foreground(style.MemorialColor())
	}

	return s
}
