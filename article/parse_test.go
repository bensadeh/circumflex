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
	require.Len(t, spans, 5)
	assert.Equal(t, formatItalic, spans[1].format)
	assert.Equal(t, "italic", spans[1].text)
	assert.Equal(t, formatCode, spans[3].format)
	assert.Equal(t, "code", spans[3].text)
}

func TestParseBlocks_LinkKeepsSurroundingSpaces(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<p>this<a href="x">recent</a> change and <a href="y">another one</a>.</p>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, "thisrecent change and another one.", blocks[0].plainText())
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
	assert.Equal(t, "func main() {\n\tfmt.Println(\"hi\")\n}", blocks[0].text)
}

func TestParseBlocks_Blockquote(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, "<blockquote><p>first line</p><p>second line</p></blockquote>")

	require.Len(t, blocks, 1)
	assert.Equal(t, blockQuote, blocks[0].kind)
	assert.Equal(t, "first line\nsecond line", blocks[0].plainText())
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
