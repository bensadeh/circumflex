package scrollbar

import (
	"strings"

	"github.com/bensadeh/circumflex/style"

	xansi "github.com/charmbracelet/x/ansi"
)

// Width is the number of columns the scrollbar occupies. Callers that render
// content up to the screen edge must reserve this much room so the bar doesn't
// overlap a glyph; Attach itself uses it to place the bar.
const Width = 1

// A slim, right-aligned bar: a thin track with a half-width thumb. The two
// quadrant glyphs let the thumb's edges land on a half-row boundary, doubling
// the apparent resolution while staying vertically symmetric.
const (
	track      = "▕" // right one-eighth block
	full       = "▐" // right half block
	topHalf    = "▝" // upper-right quadrant
	bottomHalf = "▗" // lower-right quadrant
)

// Thumb sizing, in half-cell units. A floor keeps the handle grabbable on huge
// documents; a ceiling keeps a cell of travel so a barely-overflowing view
// still visibly scrolls instead of the thumb filling the whole track.
const (
	halves       = 2
	minThumbCell = 1
)

// Attach overlays a vertical scrollbar in the rightmost column of viewportView.
// Each row is padded (or truncated) to leave room for the bar so it sits flush
// against the right edge, regardless of how wide the underlying content is. The
// thumb reflects offset within contentLines of content shown through a
// height-row viewport; when everything fits, the column stays blank.
func Attach(viewportView string, width, contentLines, height, offset int) string {
	if height <= 0 {
		return viewportView
	}

	bar := barColumn(contentLines, height, offset)
	vpLines := strings.Split(viewportView, "\n")
	colWidth := max(0, width-Width)

	var b strings.Builder
	b.Grow(height * (width + len(full)))

	for i := range height {
		if i > 0 {
			b.WriteByte('\n')
		}

		var line string
		if i < len(vpLines) {
			line = vpLines[i]
		}

		visible := xansi.StringWidth(line)
		switch {
		case visible < colWidth:
			line += strings.Repeat(" ", colWidth-visible)
		case visible > colWidth:
			line = xansi.Truncate(line, colWidth, "")
		}

		b.WriteString(line + bar[i])
	}

	return b.String()
}

// barColumn returns height single-cell rows: a track with a thumb whose size
// and position reflect the scroll state, or blank rows when all content fits.
//
// The math runs in half-cell units so the thumb's edges can land on a half-row
// boundary, rendered with the quadrant glyphs. This doubles the apparent
// resolution, letting the thumb move in half-steps rather than whole rows.
func barColumn(contentLines, height, offset int) []string {
	column := make([]string, height)

	if contentLines <= height {
		for i := range column {
			column[i] = " "
		}

		return column
	}

	units := height * halves
	minThumb := minThumbCell * halves
	maxThumb := max(minThumb, units-halves)
	thumbSize := min(max(units*height/contentLines, minThumb), maxThumb)

	travel := units - thumbSize
	maxOffset := contentLines - height

	thumbPos := 0
	if maxOffset > 0 {
		thumbPos = min(travel, offset*travel/maxOffset)
	}

	covered := func(unit int) bool {
		return unit >= thumbPos && unit < thumbPos+thumbSize
	}

	for i := range column {
		top := covered(halves * i)
		bottom := covered(halves*i + 1)

		switch {
		case top && bottom:
			column[i] = full
		case top:
			column[i] = topHalf
		case bottom:
			column[i] = bottomHalf
		default:
			column[i] = faintTrack
		}
	}

	return column
}

var faintTrack = style.Faint(track)
