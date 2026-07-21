package style

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bensadeh/circumflex/ansi"
)

func TestForegroundCode(t *testing.T) {
	t.Parallel()

	code := ForegroundCode(lipgloss.Red)

	require.NotEmpty(t, code)
	assert.True(t, strings.HasPrefix(code, "\x1b["), "should be a raw ANSI escape")
	assert.True(t, strings.HasSuffix(code, "m"), "should be a complete SGR sequence")

	assert.Empty(t, ForegroundCode(lipgloss.NoColor{}))
}

func TestWrapWithin_CapsBreakpointOvershoot(t *testing.T) {
	t.Parallel()

	// lipgloss.Wrap glues a trailing " -- " run onto a full line, overshooting
	// the width; WrapWithin must never leave a line wider than the target.
	const text = "- With increasing inference as % of total compute, if labs create efficient models -- which they can"

	for width := 10; width <= 120; width++ {
		wrapped := WrapWithin(text, width)
		for line := range strings.SplitSeq(wrapped, "\n") {
			assert.LessOrEqualf(t, lipgloss.Width(line), width,
				"width=%d left an over-wide line %q", width, line)
		}
	}
}

func TestWrapWithin_BoxStaysWithinColumn(t *testing.T) {
	t.Parallel()

	// A code box built from WrapWithin content must never grow past the column
	// it was given, even when the content contains a boundary-straddling " -- ".
	const text = "efficient models -- which they can create highly optimized models amortized over"

	for width := 12; width <= 90; width++ {
		box := ansi.Strip(RoundedBox(WrapWithin(text, width-RoundedBoxChrome), width, ""))
		for line := range strings.SplitSeq(box, "\n") {
			assert.LessOrEqualf(t, lipgloss.Width(line), width,
				"width=%d box line overflowed: %q", width, line)
		}
	}
}

func TestRoundedBox_Label(t *testing.T) {
	t.Parallel()

	unlabeled := ansi.Strip(RoundedBox("some code", 17, ""))
	assert.Equal(t, "╭───────────────╮", strings.Split(unlabeled, "\n")[0], "no label keeps the plain rule")

	labeled := ansi.Strip(RoundedBox("some code", 17, "Go"))
	assert.Equal(t, "╭────────── Go ─╮", strings.Split(labeled, "\n")[0], "label right-aligns and the rule spans the same width")

	tooNarrow := ansi.Strip(RoundedBox("wide", 0, "VeryLongLanguageName"))
	assert.Equal(t, "╭──────╮", strings.Split(tooNarrow, "\n")[0], "an oversized label drops instead of breaking the frame")
}
