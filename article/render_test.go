package article

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/scrollbar"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	showImages = ImageOptions{Show: true}
	hideImages = ImageOptions{}
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

	b := &block{kind: blockImage, spans: []span{{text: "a caption"}}}
	rendered := ansi.Strip(renderImage(b, 80, showImages))

	assert.Equal(t, "  ●●● Image a caption", rendered)
}

func TestRenderImage_UncaptionedDecorationDisappears(t *testing.T) {
	t.Parallel()

	b := &block{kind: blockImage, decorative: true}

	assert.Empty(t, renderImage(b, 80, showImages))
	assert.Empty(t, renderImage(b, 80, hideImages))
}

func TestRenderImage_CaptionedDecorationKeepsLabel(t *testing.T) {
	t.Parallel()

	b := &block{kind: blockImage, decorative: true, spans: []span{{text: "a caption"}}}

	assert.Equal(t, "  ●●● Image a caption", ansi.Strip(renderImage(b, 80, showImages)))
}

func TestRenderImageArt_HalfBlockGrid(t *testing.T) {
	t.Parallel()

	// 8x8 image: top half red, bottom half blue.
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))

	for y := range 8 {
		for x := range 8 {
			if y < 4 {
				img.Set(x, y, color.RGBA{255, 0, 0, 255})
			} else {
				img.Set(x, y, color.RGBA{0, 0, 255, 255})
			}
		}
	}

	art := renderImageArt(img, 0, 4, nil)
	lines := strings.Split(art, "\n")

	// Square image at 4 cols -> 4 pixel rows -> 2 text rows.
	require.Len(t, lines, 2)

	for _, line := range lines {
		assert.Equal(t, 4, strings.Count(line, "▀"), "each row is cols wide")
	}

	// The top text row spans the red band: red foreground over red background.
	assert.Contains(t, lines[0], "\x1b[38;2;255;0;0;48;2;255;0;0m")
	// The bottom text row spans the blue band.
	assert.Contains(t, lines[1], "\x1b[38;2;0;0;255;48;2;0;0;255m")
}

func TestRenderImageArt_TransparentPixelsKeepTerminalBackground(t *testing.T) {
	t.Parallel()

	// 8x8 image: the top quarter transparent, the rest red — a logo cut-out.
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))

	for y := 2; y < 8; y++ {
		for x := range 8 {
			img.Set(x, y, color.NRGBA{255, 0, 0, 255})
		}
	}

	art := renderImageArt(img, 0, 4, nil)
	lines := strings.Split(art, "\n")
	require.Len(t, lines, 2)

	// The first row pairs a transparent top pixel with an opaque bottom one:
	// a lower half-block on the default background, never a painted ▀.
	assert.Equal(t, 4, strings.Count(lines[0], "▄"))
	assert.Contains(t, lines[0], "\x1b[49;38;2;255;0;0m▄")
	assert.NotContains(t, lines[0], "▀")
	assert.NotContains(t, lines[0], "48;2;", "no background color painted over the transparent half")

	// The fully opaque second row keeps the two-pixel ▀ cells.
	assert.Equal(t, 4, strings.Count(lines[1], "▀"))
	assert.Contains(t, lines[1], "\x1b[38;2;255;0;0;48;2;255;0;0m")
}

func TestRenderImageArt_FullyTransparentRowRendersBlank(t *testing.T) {
	t.Parallel()

	// 8x8 image: top half transparent, bottom half green.
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))

	for y := 4; y < 8; y++ {
		for x := range 8 {
			img.Set(x, y, color.NRGBA{0, 255, 0, 255})
		}
	}

	art := renderImageArt(img, 0, 4, nil)
	lines := strings.Split(art, "\n")
	require.Len(t, lines, 2)

	assert.Equal(t, 4, strings.Count(lines[0], " "), "the transparent band is spaces")
	assert.NotContains(t, lines[0], "38;2;", "no color painted in the transparent band")
	assert.Equal(t, 4, strings.Count(lines[1], "▀"))
	assert.Contains(t, lines[1], "\x1b[38;2;0;255;0;48;2;0;255;0m")
}

func TestRenderImageArt_CompositesOnKnownBackground(t *testing.T) {
	t.Parallel()

	// 8x8 image: top half fully transparent, bottom half half-alpha red.
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))

	for y := 4; y < 8; y++ {
		for x := range 8 {
			img.Set(x, y, color.NRGBA{255, 0, 0, 128})
		}
	}

	art := renderImageArt(img, 0, 4, color.White)
	lines := strings.Split(art, "\n")
	require.Len(t, lines, 2)

	// Fully transparent pixels stay unpainted even with a known background,
	// so translucent terminals keep showing through.
	assert.Equal(t, 4, strings.Count(lines[0], " "))
	assert.NotContains(t, lines[0], "38;2;")

	// Half-alpha red over white blends to pink instead of clamping to a
	// hard opaque red or dropping the pixel.
	assert.Contains(t, lines[1], "\x1b[38;2;255;127;127;48;2;255;127;127m▀")
}

func TestRenderImage_UsesArtWhenDecoded(t *testing.T) {
	t.Parallel()

	b := &block{kind: blockImage, img: solidImage(), spans: []span{{text: "a caption"}}}

	rendered := renderImage(b, 80, showImages)

	assert.Contains(t, rendered, "▀", "renders half-blocks, not the label")
	assert.NotContains(t, rendered, imageCircle, "the ●●● label is replaced by the image")
	assert.Contains(t, ansi.Strip(rendered), "a caption", "caption stays under the image")
}

func TestRenderImage_CentersScaledDownArt(t *testing.T) {
	t.Parallel()

	const width = 80

	b := &block{kind: blockImage, img: solidImage(), spans: []span{{text: "a caption"}}}

	lines := strings.Split(ansi.Strip(renderImage(b, width, showImages)), "\n")

	inner := width - len(blockIndent)
	artCols := imageCols(0, 32, inner)
	artPad := len(blockIndent) + (inner-artCols)/2
	captionPad := len(blockIndent) + (inner-len("a caption"))/2

	require.Greater(t, len(lines), 1)

	for _, line := range lines[:len(lines)-1] {
		assert.Equal(t, strings.Repeat(" ", artPad)+strings.Repeat("▀", artCols), line)
	}

	assert.Equal(t, strings.Repeat(" ", captionPad)+"a caption", lines[len(lines)-1])
}

func TestRenderImage_FallsBackToLabelWhenHidden(t *testing.T) {
	t.Parallel()

	b := &block{kind: blockImage, img: solidImage(), spans: []span{{text: "a caption"}}}

	rendered := ansi.Strip(renderImage(b, 80, hideImages))

	assert.Equal(t, "  ●●● Image a caption", rendered, "a decoded image still shows the label when images are hidden")
	assert.NotContains(t, rendered, "▀")
}

func TestImageCols_ScalesToDisplaySize(t *testing.T) {
	t.Parallel()

	const avail = 70

	// A 60px author thumbnail (the AP byline case) collapses to the small floor
	// instead of filling the column.
	assert.Equal(t, minImageCols, imageCols(60, 704, avail))

	// A half-column image scales proportionally.
	assert.Equal(t, avail*320/referenceColumnPx, imageCols(320, 900, avail))

	// A column-width or larger image fills the available width.
	assert.Equal(t, avail, imageCols(referenceColumnPx, 1280, avail))
	assert.Equal(t, avail, imageCols(2000, 4000, avail))

	// With no display width, the intrinsic size drives the scale.
	assert.Equal(t, avail, imageCols(0, 4000, avail))
	assert.Equal(t, avail*320/referenceColumnPx, imageCols(0, 320, avail))

	// Never smaller than the visibility floor.
	assert.Equal(t, minImageCols, imageCols(1, 1, avail))
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

func TestRenderWithHeader_FullWidthLineClearsScrollbar(t *testing.T) {
	t.Parallel()

	const screenWidth = 80

	parsed := NewParsedFromHTML("<pre><code>" + strings.Repeat("x", 200) + "</code></pre>")
	rendered, _ := parsed.RenderWithHeader(72, screenWidth, "", showImages)
	out := ansi.Strip(rendered)

	widest := 0
	for line := range strings.SplitSeq(out, "\n") {
		widest = max(widest, len([]rune(line)))
	}

	assert.LessOrEqual(t, widest, screenWidth-scrollbar.Width,
		"the widest code line must leave the scrollbar column free")
	assert.Greater(t, widest, 72, "code still breaks out past the reading column")
}

func TestRenderBlocks_ImageToggleKeepsBlockCount(t *testing.T) {
	t.Parallel()

	blocks := []block{
		{kind: blockParagraph, spans: []span{{text: "before"}}},
		{kind: blockImage, img: solidImage()},
		{kind: blockParagraph, spans: []span{{text: "after"}}},
	}

	shown := renderParts(blocks, 80, 80, showImages)
	hidden := renderParts(blocks, 80, 80, hideImages)

	// Scroll re-anchoring maps block starts across a toggle, which relies on
	// both renders producing the same block sequence.
	require.Len(t, hidden, len(shown))

	assert.Contains(t, renderBlocks(blocks, 80, 80, showImages), "▀")
	assert.NotContains(t, renderBlocks(blocks, 80, 80, hideImages), "▀")
}

func TestRenderWithHeader_ReturnsBlockStarts(t *testing.T) {
	t.Parallel()

	parsed := NewParsedFromHTML("<h1>Title</h1><p>first paragraph</p><p>second paragraph</p>")
	rendered, starts := parsed.RenderWithHeader(72, 0, "meta line one\nmeta line two\n", showImages)

	lines := strings.Split(ansi.Strip(rendered), "\n")

	require.Len(t, starts, 3)
	assert.Contains(t, lines[starts[0]], "Title", "first block starts below the two header lines")
	assert.Contains(t, lines[starts[1]], "first paragraph")
	assert.Contains(t, lines[starts[2]], "second paragraph")
}

func TestRenderBlock_TableExtendsToCodeWidth(t *testing.T) {
	t.Parallel()

	b := block{kind: blockTable, rows: [][]string{
		{"Platform", "Binary"},
		{"macOS Apple Silicon", "officecli-mac-arm64-very-long-name"},
	}}

	narrow := ansi.Strip(renderBlock(&b, 20, 20, showImages))
	wide := ansi.Strip(renderBlock(&b, 20, 80, showImages))

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

	assert.Equal(t, "first\n\nsecond", renderBlocks(blocks, 80, 80, showImages))
}

func TestRenderBlocks_CodeExtendsToScreenWidth(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("x", 100)
	blocks := []block{
		{kind: blockParagraph, spans: []span{{text: strings.Repeat("word ", 30)}}},
		{kind: blockCode, text: long},
	}

	for line := range strings.SplitSeq(ansi.Strip(renderBlocks(blocks, 40, 120, showImages)), "\n") {
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

func TestRenderSpans_Bold(t *testing.T) {
	t.Parallel()

	spans := []span{{text: "loud", format: formatBold}}

	assert.Contains(t, renderSpans(spans, false), ansi.Bold+"loud"+ansi.NormalIntensity)
	assert.Contains(t, renderSpans(spans, true), ansi.Bold+"loud"+ansi.NormalIntensity+ansi.Faint,
		"inside a quote the faint must be re-opened, since SGR 22 clears it")
}

func TestRenderSpans_Underline(t *testing.T) {
	t.Parallel()

	spans := []span{{text: "marked", format: formatUnderline}}

	assert.Contains(t, renderSpans(spans, false), ansi.Underline+"marked"+ansi.UnderlineOff)
}

func TestRenderSpans_Hyperlink(t *testing.T) {
	t.Parallel()

	spans := []span{{text: "click here", href: "https://example.com"}}
	out := renderSpans(spans, false)

	assert.Contains(t, out, "8;;https://example.com", "anchor text should carry an OSC 8 hyperlink")
	assert.Contains(t, out, "\x1b[4;34m", "link text should be underlined in the default theme's blue")
	assert.Equal(t, "click here", ansi.Strip(out), "hyperlink must not change the visible text")
}

func TestRenderParagraph_StylingSurvivesWrapping(t *testing.T) {
	t.Parallel()

	spans := []span{{text: strings.Repeat("bold words here ", 5), format: formatBold}}

	for line := range strings.SplitSeq(renderParagraph(spans, 20), "\n") {
		assert.True(t, strings.HasPrefix(line, ansi.Bold), "every wrapped line must re-open bold: %q", line)
	}
}

func TestRenderParagraph_HyperlinkSurvivesWrapping(t *testing.T) {
	t.Parallel()

	spans := []span{{text: strings.Repeat("linked words here ", 5), href: "https://example.com"}}

	lines := strings.Split(renderParagraph(spans, 20), "\n")
	require.Greater(t, len(lines), 1, "the link must be long enough to wrap")

	for _, line := range lines {
		assert.Contains(t, line, "8;;https://example.com", "every wrapped line must re-open the hyperlink: %q", line)
		assert.GreaterOrEqual(t, strings.Count(line, "\x1b]8;;"), 2,
			"every wrapped line must also close the hyperlink: %q", line)
	}
}

// solidImage returns a 32×32 image filled with one opaque color.
func solidImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))

	for y := range 32 {
		for x := range 32 {
			img.Set(x, y, color.RGBA{128, 64, 32, 255})
		}
	}

	return img
}
