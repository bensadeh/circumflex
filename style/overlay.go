package style

import (
	"image/color"
	"regexp"
	"strings"

	"github.com/bensadeh/circumflex/ansi"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

var sgrSequence = regexp.MustCompile("\x1b\\[[0-9;:]*m")

// overlaySGR is the escape pair an overlay paints a span with: on opens the
// highlight, off returns the state on and the after-segment's replay can't
// restore (the forced colors) to the terminal default.
type overlaySGR struct {
	on  string
	off string
}

// Theme-dependent search highlight pairs — rebuilt by rebuildThemeStyles
// whenever the theme changes.
var (
	searchMatchSGR   overlaySGR
	searchCurrentSGR overlaySGR
)

// Both search tiers clear reverse video and intensity first, so a highlight
// reads the same on any content — inside the focused comment header
// (reverse) or faint text included. NoColor (an empty theme value) falls
// back to plain reverse video.

// matchOverlaySGR builds the all-matches highlight: the matched text
// recolored in the theme color, surroundings untouched.
func matchOverlaySGR(c color.Color) overlaySGR {
	if _, noColor := c.(lipgloss.NoColor); noColor {
		return overlaySGR{on: ansi.Reverse, off: ansi.ReverseOff}
	}

	return overlaySGR{
		on:  ansi.ReverseOff + ansi.NormalIntensity + xansi.Style{}.ForegroundColor(c).String(),
		off: ansi.DefaultForeground,
	}
}

// currentOverlaySGR builds the current-match highlight: a solid block of the
// theme color with contrasting text on top.
func currentOverlaySGR(c color.Color) overlaySGR {
	if _, noColor := c.(lipgloss.NoColor); noColor {
		return overlaySGR{on: ansi.Reverse, off: ansi.ReverseOff}
	}

	colors := xansi.Style{}.BackgroundColor(c).ForegroundColor(contrastFg(c)).String()

	return overlaySGR{
		on:  ansi.ReverseOff + ansi.NormalIntensity + colors,
		off: ansi.DefaultBackground + ansi.DefaultForeground,
	}
}

// contrastFg picks black or bright white text for a background color. The
// threshold sits below mid-gray so that palette yellow — #808000 in the
// standard table, luma ≈ 0.44 — takes black text, as it would in vim.
func contrastFg(c color.Color) color.Color {
	r, g, b, _ := c.RGBA()

	if (299*r+587*g+114*b)/1000 > 110*257 {
		return lipgloss.Black
	}

	return lipgloss.BrightWhite
}

// linkSelectSGR paints the reader's selected link: black text on blue, the
// selection block staying in the link family's hue. Fixed rather than
// theme-derived — links render in the same ANSI blue on every theme.
var linkSelectSGR = overlaySGR{
	on:  ansi.ReverseOff + ansi.NormalIntensity + xansi.Style{}.BackgroundColor(lipgloss.Blue).ForegroundColor(lipgloss.Black).String(),
	off: ansi.DefaultBackground + ansi.DefaultForeground,
}

// OverlayLinkSpans repaints the given spans of a single already-styled line
// in the link-selection colors, under the same span rules as
// OverlaySearchSpans.
func OverlayLinkSpans(line string, spans []SearchSpan) string {
	return overlaySpans(line, spans, linkSelectSGR, linkSelectSGR)
}

// linkMutedSGR marks a selected link the reader will not open — the same
// muted bar the dual-pane list draws under the open story: bright-black
// background, plain default-colored text (the link's own color and underline
// cleared).
var linkMutedSGR = overlaySGR{
	on:  ansi.ReverseOff + ansi.NormalIntensity + ansi.UnderlineOff + ansi.DefaultForeground + ansi.BgBrightBlack,
	off: ansi.DefaultBackground + ansi.DefaultForeground,
}

// OverlayMutedLinkSpans is OverlayLinkSpans in the muted colors.
func OverlayMutedLinkSpans(line string, spans []SearchSpan) string {
	return overlaySpans(line, spans, linkMutedSGR, linkMutedSGR)
}

// SearchSpan is one search hit on a line: a cell span, painted in the
// current-match colors when Current is set.
type SearchSpan struct {
	StartCell int
	EndCell   int
	Current   bool
}

// OverlaySearchSpans repaints the given spans of a single already-styled
// line in the theme's search colors, leaving the styling around and between
// them intact. Spans must be sorted and non-overlapping; offsets are display
// cells, out-of-range values clamp to the line, and empty spans drop out.
// The line is walked once however many spans it carries — one call per span
// would re-cut the previous call's output, whose replayed escapes make the
// line grow quadratically.
func OverlaySearchSpans(line string, spans []SearchSpan) string {
	return overlaySpans(line, spans, searchMatchSGR, searchCurrentSGR)
}

func overlaySpans(line string, spans []SearchSpan, match, current overlaySGR) string {
	width := xansi.StringWidth(line)

	var b strings.Builder

	pos := 0

	for _, sp := range spans {
		start := max(pos, sp.StartCell)
		end := min(sp.EndCell, width)

		if start >= end {
			continue
		}

		sgr := match
		if sp.Current {
			sgr = current
		}

		b.WriteString(xansi.Cut(line, pos, start))
		b.WriteString(paintSpan(xansi.Cut(line, start, end), sgr))

		pos = end
	}

	if pos == 0 {
		return line
	}

	b.WriteString(xansi.Cut(line, pos, width))

	return b.String()
}

// paintSpan opens the highlight over one already-cut segment. Cut replays
// every escape sequence from outside the kept range, so the segment opens
// with a replay of the line's earlier sequences — any reset among them would
// cancel the highlight. Re-asserting the highlight after each SGR keeps the
// span painted through the replay and through resets inside the span itself;
// the following segment's own replay then restores the original state.
func paintSpan(span string, sgr overlaySGR) string {
	return sgr.on + sgrSequence.ReplaceAllString(span, "${0}"+sgr.on) + sgr.off
}
