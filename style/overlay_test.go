package style

import (
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/bensadeh/circumflex/ansi"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

var sgrParams = regexp.MustCompile("^\x1b\\[([0-9;:]*)m")

// Default-theme highlight sequences: yellow text for matches, black on a
// yellow background for the current match, reverse/intensity cleared on both.
const (
	matchOn    = ansi.ReverseOff + ansi.NormalIntensity + "\x1b[33m"
	matchOff   = ansi.DefaultForeground
	currentOn  = ansi.ReverseOff + ansi.NormalIntensity + "\x1b[43;30m"
	currentOff = ansi.DefaultBackground + ansi.DefaultForeground
)

var reversePair = overlaySGR{on: ansi.Reverse, off: ansi.ReverseOff}

func overlayOne(line string, start, end int, current bool) string {
	return OverlaySearchSpans(line, []SearchSpan{{StartCell: start, EndCell: end, Current: current}})
}

// reverseCells interprets s as a terminal would, mapping the starting cell of
// each printable rune to whether reverse video is active when it is drawn.
// Only SGR sequences are understood, so inputs must not contain OSC or other
// escapes.
func reverseCells(s string) map[int]bool {
	cells := make(map[int]bool)
	reverse := false
	cell := 0

	for len(s) > 0 {
		if m := sgrParams.FindStringSubmatch(s); m != nil {
			for p := range strings.SplitSeq(m[1], ";") {
				switch p {
				case "", "0", "27":
					reverse = false
				case "7":
					reverse = true
				}
			}

			s = s[len(m[0]):]

			continue
		}

		r, size := utf8.DecodeRuneInString(s)
		cells[cell] = reverse
		cell += xansi.StringWidth(string(r))
		s = s[size:]
	}

	return cells
}

// cellState is the color state a cell is drawn with: raw SGR parameters,
// empty meaning the terminal default.
type cellState struct {
	fg string
	bg string
}

func applySGRParams(state *cellState, params string) {
	parts := strings.Split(params, ";")

	for i := 0; i < len(parts); i++ {
		switch p := parts[i]; p {
		case "", "0":
			*state = cellState{}
		case "39":
			state.fg = ""
		case "49":
			state.bg = ""
		case "38", "48":
			consumed := 1
			if i+1 < len(parts) && parts[i+1] == "5" {
				consumed = 2
			} else if i+1 < len(parts) && parts[i+1] == "2" {
				consumed = 4
			}

			end := min(i+consumed+1, len(parts))
			value := strings.Join(parts[i:end], ";")

			if p == "38" {
				state.fg = value
			} else {
				state.bg = value
			}

			i = end - 1
		default:
			switch {
			case len(p) == 2 && (p[0] == '3' || p[0] == '9'):
				state.fg = p
			case len(p) == 2 && p[0] == '4', len(p) == 3 && strings.HasPrefix(p, "10"):
				state.bg = p
			}
		}
	}
}

// colorCells interprets s as a terminal would, mapping the starting cell of
// each printable rune to the color state it is drawn with. Same SGR-only
// limitation as reverseCells.
func colorCells(s string) map[int]cellState {
	cells := make(map[int]cellState)

	var state cellState

	cell := 0

	for len(s) > 0 {
		if m := sgrParams.FindStringSubmatch(s); m != nil {
			applySGRParams(&state, m[1])
			s = s[len(m[0]):]

			continue
		}

		r, size := utf8.DecodeRuneInString(s)
		cells[cell] = state
		cell += xansi.StringWidth(string(r))
		s = s[size:]
	}

	return cells
}

// stateOf reads the color state a bare SGR string establishes.
func stateOf(sgr string) cellState {
	return colorCells(sgr + "x")[0]
}

// assertSpanReversed checks the two properties that define the reverse-video
// fallback overlay: the visible text is unchanged, and a cell renders
// reversed exactly when it was already reversed or lies inside the span.
func assertSpanReversed(t *testing.T, line string, start, end int) {
	t.Helper()

	span := []SearchSpan{{StartCell: start, EndCell: end}}
	result := overlaySpans(line, span, reversePair, reversePair)

	assert.Equal(t, ansi.Strip(line), ansi.Strip(result), "visible text must not change")

	original := reverseCells(line)

	for cell, on := range reverseCells(result) {
		want := original[cell] || (cell >= start && cell < end)
		assert.Equal(t, want, on, "reverse state at cell %d", cell)
	}
}

// assertSpansHighlighted checks the properties that define the search
// overlay: the visible text is unchanged, cells inside a span render in
// exactly that tier's colors, and cells outside keep their original state.
func assertSpansHighlighted(t *testing.T, line string, spans []SearchSpan) {
	t.Helper()

	result := OverlaySearchSpans(line, spans)

	assert.Equal(t, ansi.Strip(line), ansi.Strip(result), "visible text must not change")

	width := xansi.StringWidth(ansi.Strip(line))
	wantMatch := stateOf(matchOn)
	wantCurrent := stateOf(currentOn)
	original := colorCells(line)

	expected := func(cell int) (cellState, bool) {
		for _, sp := range spans {
			if cell >= max(0, sp.StartCell) && cell < min(sp.EndCell, width) {
				if sp.Current {
					return wantCurrent, true
				}

				return wantMatch, true
			}
		}

		return cellState{}, false
	}

	for cell, state := range colorCells(result) {
		if want, in := expected(cell); in {
			assert.Equal(t, want, state, "highlight colors at cell %d", cell)
		} else {
			assert.Equal(t, original[cell], state, "original colors at cell %d", cell)
		}
	}
}

func assertSpanHighlighted(t *testing.T, current bool, line string, start, end int) {
	t.Helper()

	assertSpansHighlighted(t, line, []SearchSpan{{StartCell: start, EndCell: end, Current: current}})
}

func TestOverlaySearchPlainLine(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "he"+matchOn+"ll"+matchOff+"o", overlayOne("hello", 2, 4, false))
	assert.Equal(t, "he"+currentOn+"ll"+currentOff+"o", overlayOne("hello", 2, 4, true))
	assert.Equal(t, matchOn+"hello"+matchOff, overlayOne("hello", 0, 5, false))
}

func TestOverlaySearchSpansSingleLinePass(t *testing.T) {
	t.Parallel()

	got := OverlaySearchSpans("needle and needle again", []SearchSpan{
		{StartCell: 0, EndCell: 6},
		{StartCell: 11, EndCell: 17, Current: true},
	})
	want := matchOn + "needle" + matchOff + " and " + currentOn + "needle" + currentOff + " again"
	assert.Equal(t, want, got)

	got = OverlaySearchSpans("aabb", []SearchSpan{
		{StartCell: 0, EndCell: 2},
		{StartCell: 2, EndCell: 4, Current: true},
	})
	assert.Equal(t, matchOn+"aa"+matchOff+currentOn+"bb"+currentOff, got, "adjacent spans keep their own tiers")
}

func TestOverlaySearchSpansManyOnStyledLine(t *testing.T) {
	t.Parallel()

	line := ansi.Red + "foo" + ansi.Reset + "bar" + ansi.Faint + "baz" + ansi.Reset + "qux"

	assertSpansHighlighted(t, line, []SearchSpan{
		{StartCell: 1, EndCell: 2},
		{StartCell: 4, EndCell: 6, Current: true},
		{StartCell: 7, EndCell: 9},
		{StartCell: 10, EndCell: 11},
	})
}

func TestOverlaySpanClamps(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "hello", overlayOne("hello", 3, 3, false), "empty span")
	assert.Equal(t, "hello", overlayOne("hello", 4, 2, false), "inverted span")
	assert.Equal(t, "hello", overlayOne("hello", 5, 9, false), "span past the end")
	assert.Equal(t, "hello", OverlaySearchSpans("hello", nil), "no spans")
	assert.Empty(t, overlayOne("", 0, 3, false), "empty line")

	assertSpanReversed(t, "hello", -2, 3)
	assertSpanReversed(t, "hello", 3, 42)
	assertSpanHighlighted(t, false, "hello", -2, 3)
	assertSpanHighlighted(t, true, "hello", 3, 42)
}

// A reset earlier in the line must not cancel the highlight: Cut replays the
// preceding sequences (reset included) at the start of the span segment.
func TestOverlaySpanAfterReset(t *testing.T) {
	t.Parallel()

	line := ansi.Red + "foo" + ansi.Reset + "bar"

	assertSpanReversed(t, line, 3, 6)
	assertSpanReversed(t, line, 4, 5)
	assertSpanHighlighted(t, false, line, 3, 6)
	assertSpanHighlighted(t, true, line, 4, 5)
}

func TestOverlaySpanAcrossStyleBoundary(t *testing.T) {
	t.Parallel()

	line := "ab" + ansi.Faint + "cd" + ansi.Reset + "ef"

	assertSpanReversed(t, line, 1, 5)
	assertSpanHighlighted(t, false, line, 1, 5)
	assertSpanHighlighted(t, false, line, 0, 6)
}

func TestOverlaySearchKeepsSurroundingColor(t *testing.T) {
	t.Parallel()

	line := ansi.Red + "hello" + ansi.Reset

	assertSpanHighlighted(t, false, line, 1, 3)
	assert.Contains(t, overlayOne(line, 1, 3, false), ansi.Red, "the surrounding color survives")
}

// The highlight must stay legible on reversed content: the span clears
// reverse video so the forced colors render as written, focused comment
// headers included.
func TestOverlaySearchOnReversedLine(t *testing.T) {
	t.Parallel()

	line := ansi.Reverse + "abc" + ansi.Reset + "def"

	assertSpanReversed(t, line, 3, 6)
	assertSpanHighlighted(t, false, line, 1, 2)

	reversed := reverseCells(overlayOne(line, 1, 2, false))
	assert.False(t, reversed[1], "the span renders with reverse cleared")
	assert.True(t, reversed[0], "outside the span the reverse survives")
	assert.True(t, reversed[2], "the replay restores reverse after the span")
}

func TestOverlaySpanWideRunes(t *testing.T) {
	t.Parallel()

	assertSpanReversed(t, "ＡＢＣＤ", 2, 6)
	assertSpanReversed(t, "aＸbＹc", 1, 4)
	assertSpanHighlighted(t, false, "ＡＢＣＤ", 2, 6)
	assertSpanHighlighted(t, true, "aＸbＹc", 1, 4)
}

func TestOverlaySpanHyperlink(t *testing.T) {
	t.Parallel()

	line := ansi.Hyperlink("https://example.com", "link") + " tail"
	result := overlayOne(line, 0, 4, false)

	assert.Equal(t, ansi.Strip(line), ansi.Strip(result))
	assert.Contains(t, result, "https://example.com", "hyperlink target must survive")
	assert.Contains(t, result, matchOn)
}

func TestSearchOverlaySGRFallsBackToReverse(t *testing.T) {
	t.Parallel()

	assert.Equal(t, reversePair, matchOverlaySGR(lipgloss.NoColor{}))
	assert.Equal(t, reversePair, currentOverlaySGR(lipgloss.NoColor{}))
}

func TestContrastFg(t *testing.T) {
	t.Parallel()

	assert.Equal(t, lipgloss.Black, contrastFg(lipgloss.Yellow), "palette yellow (#808000) takes black text")
	assert.Equal(t, lipgloss.Black, contrastFg(lipgloss.ANSIColor(214)), "orange takes black text")
	assert.Equal(t, lipgloss.BrightWhite, contrastFg(lipgloss.Blue))
	assert.Equal(t, lipgloss.BrightWhite, contrastFg(lipgloss.Color("#1a1a2e")))
}
