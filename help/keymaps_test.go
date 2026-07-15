package help

import (
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/ansi"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestKeymaps(t *testing.T) {
	t.Parallel()

	keys := new(keyList)

	first := keys.addSection("First")
	first.addKey("x", "Add item")
	first.addKey("y", "Delete item")
	first.addKey("xyz", "Other thing")
	first.addKey("a + b", "Combined")

	second := keys.addSection("Second")
	second.addKey("x", "Down")
	second.addKey("y", "Up")

	actual := keys.print(80)
	stripped := ansi.Strip(actual)

	assert.Contains(t, stripped, "─ First ─")
	assert.Contains(t, stripped, "─ Second ─")
	assert.Contains(t, stripped, "Add item")
	assert.Contains(t, stripped, "Delete item")
	assert.Contains(t, stripped, "Other thing")
	assert.Contains(t, stripped, "Combined")
	assert.Contains(t, stripped, "Down")
	assert.Contains(t, stripped, "Up")
}

// Every panel row spans exactly the width the panel is laid out at — the
// same contract meta's blocks pin in TestBlockGeometryContract, so a help
// box and the meta bar heading the same column always share a right edge.
func TestPanelsSpanExactlyTheGivenWidth(t *testing.T) {
	t.Parallel()

	keys := new(keyList)

	first := keys.addSection("First")
	first.addKey("j", "Down")
	first.addKey("k", "Up")
	first.addBreak()
	first.addKey("q", "Back")

	second := keys.addSection("Second")
	second.addKey("x", "Do the thing")

	for _, width := range []int{minPanelWidth, 59, 60, 61, 80} {
		for line := range strings.SplitSeq(keys.print(width), "\n") {
			if line == "" {
				continue
			}

			assert.Equal(t, width, xansi.StringWidth(line),
				"width %d: every panel row must reach the frame's edge exactly: %q", width, line)
		}
	}
}

func TestKeymapsNarrowWidthFallsBackToSingleColumn(t *testing.T) {
	t.Parallel()

	keys := new(keyList)
	s := keys.addSection("S")
	s.addKey("j", "Down")
	s.addKey("k", "Up")
	s.addKey("h", "Left")
	s.addKey("l", "Right")

	actual := ansi.Strip(keys.print(40))

	for _, line := range []string{"Down", "Up", "Left", "Right"} {
		assert.Contains(t, actual, line)
	}
}
