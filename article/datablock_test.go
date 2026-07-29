package article

import (
	"bytes"
	nurl "net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/net/html"
)

func flattenedHTML(t *testing.T, src string) string {
	t.Helper()

	node, err := html.Parse(strings.NewReader(src))
	require.NoError(t, err)

	normalizeDataBlocks(node)

	var buf bytes.Buffer
	require.NoError(t, html.Render(&buf, node))

	return buf.String()
}

func TestNormalizeDataBlocks_SplicesSectionContent(t *testing.T) {
	t.Parallel()

	out := flattenedHTML(t, `<div class="stack">`+
		`<div data-block="text"><div class="spacer"><div data-testid="rich-text" data-block="text">`+
		`<div class="rich-text-container"><p>First.</p><p>Second.</p></div></div></div></div>`+
		`<div data-block="image"><div class="spacer"><div class="full-width">`+
		`<figure><img src="https://example.com/a.jpg" alt="A photo"/></figure></div></div></div>`+
		`<div data-block="subheadline"><div class="spacer"><h2>Part two</h2></div></div>`+
		`</div>`)

	assert.Contains(t, out, `<div class="stack"><p>First.</p><p>Second.</p><figure>`)
	assert.Contains(t, out, `</figure><h2>Part two</h2></div>`)
	assert.NotContains(t, out, "data-block")
	assert.NotContains(t, out, "spacer")
}

// A lone data-block element is not the sectioned-article signature — a page
// reusing the attribute name for something else must pass through untouched.
func TestNormalizeDataBlocks_IgnoresLoneSection(t *testing.T) {
	t.Parallel()

	src := `<div><div data-block="text"><div><p>Alone.</p></div></div><p>Prose.</p></div>`

	assert.Contains(t, flattenedHTML(t, src), src)
}

// BBC News wraps every article section in its own chain of styled-component
// divs; readability scores each chain in isolation, keeps the biggest text
// section, and drops the rest of the article along with every figure. The
// sections must reach the parser whole.
func TestExtractReadable_DataBlockSectionsSurvive(t *testing.T) {
	t.Parallel()

	filler := `<p>Enough readable content for the extractor to accept this page as an article, ` +
		`repeated to pass its length heuristics without any help from the sections below.</p>`

	section := func(kind, inner string) string {
		return `<div data-block="` + kind + `" class="wrapper"><div class="spacer">` + inner + `</div></div>`
	}

	page := `<html><head><title>T</title></head><body><article><div class="stack">` +
		section("text", `<div data-testid="rich-text" data-block="text"><div class="container">`+
			strings.Repeat(filler, 6)+`</div></div>`) +
		section("image", `<div class="holder"><figure>`+
			`<img src="https://example.com/chart.jpg" alt="A chart"/></figure></div>`) +
		section("subheadline", `<h2>The second half</h2>`) +
		section("text", `<div data-testid="rich-text" data-block="text"><div class="container">`+
			`<p>The closing section, shorter than the opener, must still make it through.</p></div></div>`) +
		`</div></article></body></html>`

	u, err := nurl.Parse("https://www.bbc.co.uk/news/articles/example")
	require.NoError(t, err)

	node, _, err := extractReadable([]byte(page), u)
	require.NoError(t, err)

	var (
		imageURLs []string
		headings  []string
		text      strings.Builder
	)

	for _, b := range parseBlocks(node) {
		switch b.kind {
		case blockImage:
			imageURLs = append(imageURLs, b.imageURL)

		case blockHeading:
			headings = append(headings, b.text)

		case blockParagraph, blockList, blockQuote, blockCode,
			blockTable, blockDivider, blockVerbatim, blockInfobox:
			text.WriteString(b.plainText())
		}
	}

	assert.Equal(t, []string{"https://example.com/chart.jpg"}, imageURLs)
	assert.Equal(t, []string{"The second half"}, headings)
	assert.Contains(t, text.String(), "The closing section")
}
