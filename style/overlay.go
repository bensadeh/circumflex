package style

import (
	"regexp"

	"github.com/bensadeh/circumflex/ansi"

	xansi "github.com/charmbracelet/x/ansi"
)

var sgrSequence = regexp.MustCompile("\x1b\\[[0-9;:]*m")

// OverlaySpan repaints cells [startCell, endCell) of a single already-styled
// line in reverse video, leaving the styling around and inside the span
// intact. Offsets are display cells; out-of-range values clamp to the line,
// and an empty span returns the line unchanged.
func OverlaySpan(line string, startCell, endCell int) string {
	width := xansi.StringWidth(line)

	startCell = max(0, startCell)
	endCell = min(endCell, width)

	if startCell >= endCell {
		return line
	}

	before := xansi.Cut(line, 0, startCell)
	span := xansi.Cut(line, startCell, endCell)
	after := xansi.Cut(line, endCell, width)

	// Cut replays every escape sequence from outside the kept range, so the
	// span opens with a replay of the line's earlier sequences — any reset
	// among them would cancel the highlight. Re-asserting reverse after each
	// SGR keeps the span highlighted through the replay and through resets
	// inside the span itself; the trailing segment's own replay then restores
	// the original state after the span.
	span = ansi.Reverse + sgrSequence.ReplaceAllString(span, "${0}"+ansi.Reverse)

	return before + span + ansi.ReverseOff + after
}
