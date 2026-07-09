package article

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/net/html"
)

func blocksFromHTML(t *testing.T, src string) []block {
	t.Helper()

	node, err := html.Parse(strings.NewReader(src))
	require.NoError(t, err)

	return parseBlocks(node)
}

func TestParseBlocks_Paragraph(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<p>Hello \n\t world</p>")

	require.Len(t, blocks, 1)
	assert.Equal(t, blockParagraph, blocks[0].kind)
	assert.Equal(t, "Hello world", blocks[0].plainText())
}

func TestParseBlocks_InlineFormatting(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<p>plain <em>italic</em> <code>code</code> <strong>bold</strong> <a href="x">link</a> end</p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "plain italic code bold link end", blocks[0].plainText())

	spans := blocks[0].spans
	require.Len(t, spans, 7)
	assert.Equal(t, formatItalic, spans[1].format)
	assert.Equal(t, "italic", spans[1].text)
	assert.Equal(t, formatCode, spans[3].format)
	assert.Equal(t, "code", spans[3].text)
	assert.Equal(t, formatBold, spans[5].format)
	assert.Equal(t, "bold", spans[5].text)
}

func TestParseBlocks_UnderlineFormatting(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<p>plain <u>underlined</u> and <ins>inserted</ins> end</p>`)

	require.Len(t, blocks, 1)

	spans := blocks[0].spans
	require.Len(t, spans, 5)
	assert.Equal(t, formatUnderline, spans[1].format)
	assert.Equal(t, "underlined", spans[1].text)
	assert.Equal(t, formatUnderline, spans[3].format)
	assert.Equal(t, "inserted", spans[3].text)
}

func TestParseBlocks_LinkKeepsSurroundingSpaces(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<p>this<a href="x">recent</a> change and <a href="y">another one</a>.</p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "thisrecent change and another one.", blocks[0].plainText())
}

func TestParseBlocks_BrIsHardBreak(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<p>Affected Versions: <br>
		* US_FH1201<br>
		* US_W15E</p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "Affected Versions:\n* US_FH1201\n* US_W15E", blocks[0].plainText(),
		"a br must break the line, not collapse into the surrounding spaces")
}

func TestParseBlocks_BrRunsCapAtBlankLine(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<p>above<br><br><br>below</p>")

	require.Len(t, blocks, 1)
	assert.Equal(t, "above\n\nbelow", blocks[0].plainText())
}

func TestParseBlocks_BrOnlyContentIsSkipped(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<p>a</p><br><br><p>b</p>")

	require.Len(t, blocks, 2)
	assert.Equal(t, "a", blocks[0].plainText())
	assert.Equal(t, "b", blocks[1].plainText())
}

func TestParseBlocks_TableCellFlattensBr(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<table><tr><td>first<br>second</td><td>x</td></tr></table>")

	require.Len(t, blocks, 1)
	assert.Equal(t, []string{"first second", "x"}, blocks[0].rows[0],
		"rows render as single lines, so cell breaks must flatten to spaces")
}

func TestParseBlocks_HeadingLevels(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<h1>One</h1><h2>Two</h2><h3>Three</h3>")

	require.Len(t, blocks, 3)

	for i, want := range []int{1, 2, 3} {
		assert.Equal(t, blockHeading, blocks[i].kind)
		assert.Equal(t, want, blocks[i].level)
	}
}

func TestParseBlocks_HeadingNormalization(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<h2>Top</h2><p>text</p><h4>Sub</h4>")

	require.Len(t, blocks, 3)
	assert.Equal(t, 1, blocks[0].level)
	assert.Equal(t, 2, blocks[2].level)
}

func TestParseBlocks_NestedList(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<ul>
		<li>first</li>
		<li>parent
			<ul><li>nested</li></ul>
		</li>
	</ul>`)

	require.Len(t, blocks, 1)
	require.Len(t, blocks[0].items, 3)

	assert.Equal(t, 0, blocks[0].items[0].depth)
	assert.Equal(t, "parent", spanText(blocks[0].items[1].spans))
	assert.Equal(t, 1, blocks[0].items[2].depth)
	assert.Equal(t, "nested", spanText(blocks[0].items[2].spans))
}

func TestParseBlocks_OrderedListNumbering(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<ol start="3"><li>a</li><li>b</li></ol>`)

	require.Len(t, blocks, 1)
	require.Len(t, blocks[0].items, 2)
	assert.Equal(t, 3, blocks[0].items[0].number)
	assert.Equal(t, 4, blocks[0].items[1].number)
}

func TestParseBlocks_PreservesCodeVerbatim(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<pre><code>func main() {\n\tfmt.Println(\"hi\")\n}</code></pre>")

	require.Len(t, blocks, 1)
	assert.Equal(t, blockCode, blocks[0].kind)
	assert.Equal(t, "func main() {\n        fmt.Println(\"hi\")\n}", blocks[0].text,
		"tabs expand to 8 columns so the wrapper's width math matches the terminal")
}

func TestParseBlocks_Blockquote(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<blockquote><p>first line</p><p>second line</p></blockquote>")

	require.Len(t, blocks, 1)
	assert.Equal(t, blockQuote, blocks[0].kind)
	assert.Equal(t, "first line\n\nsecond line", blocks[0].plainText(),
		"quoted paragraphs keep a blank line between them, like top-level blocks")
}

func TestParseBlocks_Table(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<table>
		<caption>Numbers</caption>
		<thead><tr><th>Name</th><th>Value</th></tr></thead>
		<tbody>
			<tr><td>Foo</td><td>1</td></tr>
			<tr><td></td><td></td></tr>
		</tbody>
	</table>`)

	require.Len(t, blocks, 2)
	assert.Equal(t, blockParagraph, blocks[0].kind)
	assert.Equal(t, "Numbers", blocks[0].plainText())

	assert.Equal(t, blockTable, blocks[1].kind)
	require.Len(t, blocks[1].rows, 2, "empty rows should be skipped")
	assert.Equal(t, []string{"Name", "Value"}, blocks[1].rows[0])
}

func TestParseBlocks_FigurePrefersFigcaption(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<figure><img src="x.png" alt="alt text"><figcaption>The caption</figcaption></figure>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockImage, blocks[0].kind)
	assert.Equal(t, "The caption", blocks[0].plainText())
}

func TestParseBlocks_InlineImageBecomesOwnBlock(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<p>before <img src="x.png" alt="a chart"> after</p>`)

	require.Len(t, blocks, 2)
	assert.Equal(t, blockParagraph, blocks[0].kind)
	assert.Equal(t, "before after", blocks[0].plainText())
	assert.Equal(t, blockImage, blocks[1].kind)
	assert.Equal(t, "a chart", blocks[1].plainText())
}

func TestParseBlocks_SrcsetPrefersRightSizedVariant(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<img src="full.jpeg" srcset="a-300.jpeg 300w, a-1024.jpeg 1024w, a-768.jpeg 768w" alt="a">`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "a-768.jpeg", blocks[0].imageURL, "smallest candidate covering maxRetainedPx wins over full-size src")
}

func TestParseBlocks_SrcsetWithoutUsableWidthsFallsBackToSrc(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<img src="full.jpeg" srcset="a-300.jpeg 300w, a-2x.jpeg 2x" alt="a">`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "full.jpeg", blocks[0].imageURL, "no candidate covers maxRetainedPx, so the eager src wins")
}

func TestParseBlocks_Strikethrough(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<p>was <del>$99</del> now $79</p>`)

	require.Len(t, blocks, 1)
	spans := blocks[0].spans
	require.Len(t, spans, 3)
	assert.Equal(t, formatStrike, spans[1].format)
	assert.Equal(t, "$99", spans[1].text)
}

func TestParseBlocks_ImageAltStripsControlBytes(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<img src=\"x.png\" alt=\"cap\x1b[2Ationtext\x07\">")

	require.Len(t, blocks, 1)
	assert.Equal(t, blockImage, blocks[0].kind)
	assert.Equal(t, "captiontext", blocks[0].plainText(), "control bytes in alt must be stripped")
}

func TestParseBlocks_LinkWithControlBytesNotLinked(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"C0 escape/BEL": "https://x.com/\x1b]2;pwned\x07",
		"C1 ST/CSI":     "https://x.com/\u009c\u009b6n", // 8-bit string terminator then CSI
		"DEL":           "https://x.com/\u007f",
	}

	for name, href := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			blocks := blocksFromHTML(t, "<p><a href=\""+href+"\">click</a></p>")

			require.Len(t, blocks, 1)
			spans := blocks[0].spans
			require.Len(t, spans, 1)
			assert.Equal(t, "click", spans[0].text)
			assert.Empty(t, spans[0].href, "a href with control characters must not become a hyperlink")
		})
	}
}

func TestParseBlocks_DedupeKeepsSameTextDifferentLevel(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<h1>FAQ</h1><h2>FAQ</h2><p>body</p>")

	require.Len(t, blocks, 3, "a repeated title at a different heading level must survive dedup")
	assert.Equal(t, blockHeading, blocks[0].kind)
	assert.Equal(t, blockHeading, blocks[1].kind)
}

func TestParseBlocks_LinksKeepHref(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<p><a href="https://a.com">first</a> and <a href="https://b.com">second</a> and <a href="#fn1">footnote</a></p>`)

	require.Len(t, blocks, 1)
	spans := blocks[0].spans

	require.Len(t, spans, 4, "different hrefs must not merge; fragment link merges into plain text")
	assert.Equal(t, "https://a.com", spans[0].href)
	assert.Empty(t, spans[1].href)
	assert.Equal(t, "https://b.com", spans[2].href)
	assert.Equal(t, " and footnote", spans[3].text)
	assert.Empty(t, spans[3].href)
}

func TestParseBlocks_SuperscriptAndSubscript(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<p>E = mc<sup>2</sup> and H<sub>2</sub>O and a note<sup>[1]</sup></p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "E = mc² and H₂O and a note[1]", blocks[0].plainText(),
		"mappable runes convert, unmappable content falls back verbatim")
}

func TestParseBlocks_IgnoresComments(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<p>text</p><!--THE END-->")

	require.Len(t, blocks, 1)
	assert.Equal(t, "text", blocks[0].plainText())
}

func TestParseBlocks_ImplicitParagraphInContainer(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<div>loose text <em>styled</em><p>real paragraph</p></div>")

	require.Len(t, blocks, 2)
	assert.Equal(t, "loose text styled", blocks[0].plainText())
	assert.Equal(t, "real paragraph", blocks[1].plainText())
}

// Readability's div-to-p conversion creates elements with an empty DataAtom;
// the walker must classify them by tag name, not atom.
func TestParseBlocks_SynthesizedNodesWithoutAtoms(t *testing.T) {
	t.Parallel()

	node, err := html.Parse(strings.NewReader(
		"<div><p>assistant · Claude Fable 5</p> <p>I could reasonably skip it</p></div><ul><li>item</li></ul>"))
	require.NoError(t, err)

	for n := range node.Descendants() {
		n.DataAtom = 0
	}

	blocks := parseBlocks(node)

	require.Len(t, blocks, 3)
	assert.Equal(t, blockParagraph, blocks[0].kind)
	assert.Equal(t, "assistant · Claude Fable 5", blocks[0].plainText())
	assert.Equal(t, blockParagraph, blocks[1].kind)
	assert.Equal(t, "I could reasonably skip it", blocks[1].plainText())
	assert.Equal(t, blockList, blocks[2].kind)
}

func TestParseBlocks_BlockContentInsideCustomElement(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<markdown-accessiblity-table>
		<table><tr><th>Tool</th><th>Stars</th></tr><tr><td>OfficeCLI</td><td>1k</td></tr></table>
	</markdown-accessiblity-table>
	<custom-note>inline only</custom-note>`)

	require.Len(t, blocks, 2)
	assert.Equal(t, blockTable, blocks[0].kind)
	assert.Equal(t, []string{"Tool", "Stars"}, blocks[0].rows[0])
	assert.Equal(t, blockParagraph, blocks[1].kind)
	assert.Equal(t, "inline only", blocks[1].plainText())
}

func TestParseBlocks_Divider(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<p>a</p><hr><p>b</p>")

	require.Len(t, blocks, 3)
	assert.Equal(t, blockDivider, blocks[1].kind)
}

func TestParseBlocks_SkipsEmptyContent(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<p>  </p><h2> </h2><pre>  </pre><ul><li></li></ul>")

	assert.Empty(t, blocks)
}

func TestParseBlocks_DedupesConsecutiveDuplicates(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<img src="a.png" alt="hero"><img src="b.png" alt="hero">`+
		`<p>Credit: Getty</p><p>Credit: Getty</p><p>Credit: Getty was here twice</p>`) //nolint:dupword // duplication is what is being tested

	require.Len(t, blocks, 3)
	assert.Equal(t, blockImage, blocks[0].kind)
	assert.Equal(t, "Credit: Getty", blocks[1].plainText())
	assert.Equal(t, "Credit: Getty was here twice", blocks[2].plainText())
}

func TestParseFigure_PrefersRealImageOverPlaceholder(t *testing.T) {
	t.Parallel()

	// BBC pattern: a grey lazy-load placeholder img next to the real one.
	blocks := blocksFromHTML(t, `<figure>`+
		`<img src="https://static.example.com/grey-placeholder.png">`+
		`<img src="https://ichef.example.com/real-photo.webp">`+
		`<figcaption>A caption</figcaption></figure>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockImage, blocks[0].kind)
	assert.Equal(t, "https://ichef.example.com/real-photo.webp", blocks[0].imageURL)
	assert.Equal(t, "A caption", blocks[0].plainText())
}

func TestIsPlaceholderURL(t *testing.T) {
	t.Parallel()

	assert.True(t, isPlaceholderURL("https://static.files.bbci.co.uk/.../grey-placeholder.png"))
	assert.True(t, isPlaceholderURL("https://cdn.example.com/assets/spacer.gif"))
	assert.False(t, isPlaceholderURL("https://ichef.bbci.co.uk/images/ic/480xn/p0nvtng9.jpg.webp"))
}

func TestCollapseWhitespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"a  b", "a b"},
		{"a\n\tb", "a b"},
		{" a ", " a "},
		{"\n", " "},
		{"", ""},
		{"a\u200bb", "ab"},
		{"a\u00adb", "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, collapseWhitespace(tt.input))
		})
	}
}

func TestNormalizeSpans_MergesAndTrims(t *testing.T) {
	t.Parallel()

	spans := normalizeSpans([]span{
		{text: " a "},
		{text: " b "},
		{text: "c ", format: formatItalic},
		{text: " "},
	})

	require.Len(t, spans, 2)
	assert.Equal(t, "a b ", spans[0].text)
	assert.Equal(t, "c", spans[1].text)
}

func TestNormalizeSpans_NewlineWinsOverSpace(t *testing.T) {
	t.Parallel()

	spans := normalizeSpans([]span{
		{text: "a "},
		{text: "\n"},
		{text: " b"},
	})

	require.Len(t, spans, 1)
	assert.Equal(t, "a\nb", spans[0].text)
}

func TestNormalizeSpans_DropsEdgeNewlines(t *testing.T) {
	t.Parallel()

	spans := normalizeSpans([]span{
		{text: "\n"},
		{text: "a"},
		{text: "\n"},
	})

	require.Len(t, spans, 1)
	assert.Equal(t, "a", spans[0].text)
}
