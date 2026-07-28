package article

import (
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/ansi"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const kumamotoInfobox = `
	<table class="infobox vevent">
		<caption class="infobox-title summary">2016 Kumamoto earthquakes</caption>
		<tbody>
			<tr><td class="infobox-image">A destroyed house in Kumamoto</td></tr>
			<tr><th class="infobox-label">Magnitude</th><td class="infobox-data">7.0 Mw</td></tr>
			<tr><th class="infobox-label">Depth</th><td class="infobox-data">10&nbsp;km</td></tr>
			<tr><th class="infobox-label">Casualties</th><td class="infobox-data">277 dead, 2,809 injured</td></tr>
		</tbody>
	</table>
	<p>The 2016 Kumamoto earthquakes were a series of earthquakes.</p>`

// The caption becomes the panel title, labeled pairs become rows, and the
// unlabeled image row drops with the rest of the sidebar chrome.
func TestParseInfobox_LabeledPairs(t *testing.T) {
	t.Parallel()

	blocks := parsedBlocks(t, kumamotoInfobox)

	require.Len(t, blocks, 2)
	assert.Equal(t, blockInfobox, blocks[0].kind)
	assert.Equal(t, "2016 Kumamoto earthquakes", blocks[0].text)
	assert.Equal(t, [][]string{
		{"Magnitude", "7.0 Mw"},
		{"Depth", "10 km"},
		{"Casualties", "277 dead, 2,809 injured"},
	}, blocks[0].rows)
	assert.Equal(t, "The 2016 Kumamoto earthquakes were a series of earthquakes.", blocks[1].plainText())
}

// The class name alone is not the template: a table using "infobox" for its
// own markup but without labeled pairs stays a regular table.
func TestParseInfobox_ClassWithoutPairsStaysTable(t *testing.T) {
	t.Parallel()

	blocks := parsedBlocks(t, `
		<table class="infobox">
			<tbody><tr><td>Speed</td><td>fast</td></tr></tbody>
		</table>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockTable, blocks[0].kind)
}

func TestRenderInfobox_FramedPanel(t *testing.T) {
	t.Parallel()

	b := block{
		kind: blockInfobox,
		text: "2016 Kumamoto earthquakes",
		rows: [][]string{
			{"Depth", "10 km"},
			{"Casualties", "277 dead, 2,809 injured (including indirect deaths)"},
		},
	}

	rendered := ansi.Strip(renderBlock(&b, 40, 80, hideImages))
	lines := strings.Split(rendered, "\n")

	assert.Contains(t, lines[0], "╭")
	assert.Contains(t, lines[0], "2016 Kumamoto earthquakes")
	assert.Contains(t, rendered, "Depth")
	assert.Contains(t, rendered, "10 km")
	assert.Contains(t, lines[len(lines)-1], "╰")

	require.Greater(t, len(lines), 4, "the long value wraps onto a continuation row")

	for _, line := range lines {
		assert.LessOrEqual(t, lipgloss.Width(line), 40, "the frame spans the content width")
	}
}

func parsedBlocks(t *testing.T, src string) []block {
	t.Helper()

	return NewParsedFromHTML(src).blocks
}
