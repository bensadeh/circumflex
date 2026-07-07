package article

import (
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/ansi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderList_GlyphsByDepth(t *testing.T) {
	t.Parallel()

	items := []listItem{
		{depth: 0, spans: []span{{text: "zero"}}},
		{depth: 1, spans: []span{{text: "one"}}},
		{depth: 2, spans: []span{{text: "two"}}},
		{depth: 3, spans: []span{{text: "three"}}},
		{depth: 4, spans: []span{{text: "four"}}},
	}

	lines := strings.Split(renderList(items, 80), "\n")

	require.Len(t, lines, 5)
	assert.Equal(t, "  - zero", lines[0])
	assert.Equal(t, "    • one", lines[1])
	assert.Equal(t, "      ◦ two", lines[2])
	assert.Equal(t, "        ▪ three", lines[3])
	assert.Equal(t, "          ▫ four", lines[4])
}

func TestRenderList_NumberAlignment(t *testing.T) {
	t.Parallel()

	var items []listItem
	for i := 1; i <= 10; i++ {
		items = append(items, listItem{number: i, spans: []span{{text: "item"}}})
	}

	lines := strings.Split(renderList(items, 80), "\n")

	require.Len(t, lines, 10)
	assert.Equal(t, "   1. item", lines[0])
	assert.Equal(t, "  10. item", lines[9])
}

func TestRenderList_WrapsWithAlignedContinuation(t *testing.T) {
	t.Parallel()

	items := []listItem{{spans: []span{{text: "a list item long enough that it must wrap"}}}}

	lines := strings.Split(renderList(items, 30), "\n")

	require.Greater(t, len(lines), 1)
	assert.True(t, strings.HasPrefix(lines[0], "  - a"))
	assert.True(t, strings.HasPrefix(lines[1], "    "), "continuation should align past the bullet")
}

func TestRenderHeading_MarkerAndIndent(t *testing.T) {
	t.Parallel()

	h1 := ansi.Strip(renderHeading(1, "Top", 80))
	assert.Equal(t, "■ Top", h1)

	h3 := ansi.Strip(renderHeading(3, "Deep", 80))
	assert.Equal(t, "    ■ Deep", h3)
}

func TestRenderQuote_PrefixesIndentBar(t *testing.T) {
	t.Parallel()

	quote := renderQuote([]span{{text: "quoted words"}}, 80)

	for line := range strings.SplitSeq(ansi.Strip(quote), "\n") {
		assert.True(t, strings.HasPrefix(line, "   ▎"), "got %q", line)
	}
}

func TestRenderCode_IndentsAllLines(t *testing.T) {
	t.Parallel()

	code := ansi.Strip(renderCode("line one\nline two", 80))

	assert.Equal(t, "  line one\n  line two", code)
}

func TestRenderImage_LabelAndCaption(t *testing.T) {
	t.Parallel()

	image := ansi.Strip(renderImage([]span{{text: "a caption"}}, 80))

	assert.Equal(t, "  ●●● Image a caption", image)
}

func TestRenderTable_AlignsColumns(t *testing.T) {
	t.Parallel()

	rows := [][]string{
		{"Name", "Value"},
		{"Foo", "1"},
		{"Longer", "2"},
	}

	lines := strings.Split(ansi.Strip(renderTable(rows, 80)), "\n")

	require.Len(t, lines, 4)
	assert.Equal(t, "Name    Value", lines[0])
	assert.Equal(t, "------  -----", lines[1])
	assert.Equal(t, "Foo     1", lines[2])
	assert.Equal(t, "Longer  2", lines[3])
}

func TestRenderBlock_TableExtendsToCodeWidth(t *testing.T) {
	t.Parallel()

	b := block{kind: blockTable, rows: [][]string{
		{"Platform", "Binary"},
		{"macOS Apple Silicon", "officecli-mac-arm64-very-long-name"},
	}}

	narrow := ansi.Strip(renderBlock(&b, 20, 20))
	wide := ansi.Strip(renderBlock(&b, 20, 80))

	assert.Contains(t, narrow, "…", "at a narrow code width the table truncates")
	assert.NotContains(t, wide, "…", "with screen room the table uses the full code width")
	assert.Contains(t, wide, "officecli-mac-arm64-very-long-name")
}

func TestRenderDivider_FitsWidth(t *testing.T) {
	t.Parallel()

	divider := ansi.Strip(renderDivider(20))

	assert.Equal(t, "  "+strings.Repeat("-", 16), divider)
}

func TestRenderBlocks_JoinsWithBlankLine(t *testing.T) {
	t.Parallel()

	blocks := []block{
		{kind: blockParagraph, spans: []span{{text: "first"}}},
		{kind: blockParagraph, spans: []span{{text: "second"}}},
	}

	assert.Equal(t, "first\n\nsecond", renderBlocks(blocks, 80, 80))
}

func TestRenderBlocks_CodeExtendsToScreenWidth(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("x", 100)
	blocks := []block{
		{kind: blockParagraph, spans: []span{{text: strings.Repeat("word ", 30)}}},
		{kind: blockCode, text: long},
	}

	for line := range strings.SplitSeq(ansi.Strip(renderBlocks(blocks, 40, 120)), "\n") {
		if strings.Contains(line, "word") {
			assert.LessOrEqual(t, len(line), 40, "prose stays in the reading column")
		}

		if strings.Contains(line, "x") {
			assert.Equal(t, "  "+long, line, "code gets the full screen width")
		}
	}
}

func TestRenderSpans_ItalicInvertsInsideQuotes(t *testing.T) {
	t.Parallel()

	spans := []span{{text: "emphasis", format: formatItalic}}

	assert.Contains(t, renderSpans(spans, false), ansi.Italic+"emphasis"+ansi.ItalicOff)
	assert.Contains(t, renderSpans(spans, true), ansi.ItalicOff+"emphasis"+ansi.Italic)
}

func TestRenderSpans_Strikethrough(t *testing.T) {
	t.Parallel()

	spans := []span{{text: "$99", format: formatStrike}}

	assert.Contains(t, renderSpans(spans, false), ansi.Strikethrough+"$99"+ansi.StrikethroughOff)
}

func TestRenderSpans_Hyperlink(t *testing.T) {
	t.Parallel()

	spans := []span{{text: "click here", href: "https://example.com"}}
	out := renderSpans(spans, false)

	assert.Contains(t, out, "8;;https://example.com", "anchor text should carry an OSC 8 hyperlink")
	assert.Equal(t, "click here", ansi.Strip(out), "hyperlink must not change the visible text")
}
