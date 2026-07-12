package style

import (
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/bensadeh/circumflex/ansi"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

var sgrParams = regexp.MustCompile("^\x1b\\[([0-9;:]*)m")

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

// assertSpanReversed checks the two properties that define OverlaySpan: the
// visible text is unchanged, and a cell renders reversed exactly when it was
// already reversed or lies inside the span.
func assertSpanReversed(t *testing.T, line string, start, end int) {
	t.Helper()

	result := OverlaySpan(line, start, end)

	assert.Equal(t, ansi.Strip(line), ansi.Strip(result), "visible text must not change")

	original := reverseCells(line)

	for cell, on := range reverseCells(result) {
		want := original[cell] || (cell >= start && cell < end)
		assert.Equal(t, want, on, "reverse state at cell %d", cell)
	}
}

func TestOverlaySpanPlainLine(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "he"+ansi.Reverse+"ll"+ansi.ReverseOff+"o", OverlaySpan("hello", 2, 4))
	assert.Equal(t, ansi.Reverse+"hello"+ansi.ReverseOff, OverlaySpan("hello", 0, 5))
}

func TestOverlaySpanClamps(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "hello", OverlaySpan("hello", 3, 3), "empty span")
	assert.Equal(t, "hello", OverlaySpan("hello", 4, 2), "inverted span")
	assert.Equal(t, "hello", OverlaySpan("hello", 5, 9), "span past the end")
	assert.Empty(t, OverlaySpan("", 0, 3), "empty line")

	assertSpanReversed(t, "hello", -2, 3)
	assertSpanReversed(t, "hello", 3, 42)
}

// A reset earlier in the line must not cancel the highlight: Cut replays the
// preceding sequences (reset included) at the start of the span segment.
func TestOverlaySpanAfterReset(t *testing.T) {
	t.Parallel()

	line := ansi.Red + "foo" + ansi.Reset + "bar"

	assertSpanReversed(t, line, 3, 6)
	assertSpanReversed(t, line, 4, 5)
}

func TestOverlaySpanAcrossStyleBoundary(t *testing.T) {
	t.Parallel()

	line := "ab" + ansi.Faint + "cd" + ansi.Reset + "ef"

	assertSpanReversed(t, line, 1, 5)
	assertSpanReversed(t, line, 0, 6)
}

func TestOverlaySpanKeepsColor(t *testing.T) {
	t.Parallel()

	line := ansi.Red + "hello" + ansi.Reset

	assertSpanReversed(t, line, 1, 3)
	assert.Contains(t, OverlaySpan(line, 1, 3), ansi.Red)
}

func TestOverlaySpanOnReversedLine(t *testing.T) {
	t.Parallel()

	line := ansi.Reverse + "abc" + ansi.Reset + "def"

	assertSpanReversed(t, line, 1, 2)
	assertSpanReversed(t, line, 3, 6)
	assertSpanReversed(t, line, 2, 4)
}

func TestOverlaySpanWideRunes(t *testing.T) {
	t.Parallel()

	assertSpanReversed(t, "ＡＢＣＤ", 2, 6)
	assertSpanReversed(t, "aＸbＹc", 1, 4)
}

func TestOverlaySpanHyperlink(t *testing.T) {
	t.Parallel()

	line := ansi.Hyperlink("https://example.com", "link") + " tail"
	result := OverlaySpan(line, 0, 4)

	assert.Equal(t, ansi.Strip(line), ansi.Strip(result))
	assert.Contains(t, result, "https://example.com", "hyperlink target must survive")
	assert.Contains(t, result, ansi.Reverse)
}
