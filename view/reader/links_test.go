package reader

import (
	"testing"

	"github.com/bensadeh/circumflex/article"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spanText cuts the plain rendering at a span's cells; the fixtures are
// ASCII, so bytes and cells coincide.
func spanText(m *Model, s int, l int) string {
	span := m.links[s].spans[l]

	return m.PlainLines()[span.Line][span.StartCell:span.EndCell]
}

func TestExtractLinks_SingleLink(t *testing.T) {
	parsed := article.NewParsedFromHTML(
		`<p>Some text with <a href="https://example.com/page">a link</a> inside.</p>`)
	m := NewWithArticle(parsed, "Title", 72, 100, 30, Options{}, nil)

	require.Len(t, m.links, 1)
	assert.Equal(t, "https://example.com/page", m.links[0].url)
	require.Len(t, m.links[0].spans, 1)
	assert.Equal(t, "a link", spanText(m, 0, 0))
}

func TestExtractLinks_WrappedLinkIsOneLink(t *testing.T) {
	parsed := article.NewParsedFromHTML(
		`<p>Intro text here. <a href="https://example.com/long">a fairly long anchor text that wraps across the column</a> after.</p>`)
	m := NewWithArticle(parsed, "Title", 30, 40, 30, Options{}, nil)

	require.Len(t, m.links, 1)
	require.Greater(t, len(m.links[0].spans), 1, "the narrow column must split the anchor text")

	for i, span := range m.links[0].spans[1:] {
		assert.Equal(t, m.links[0].spans[i].Line+1, span.Line, "spans sit on consecutive lines")
	}
}

func TestExtractLinks_TwoLinksOnOneLine(t *testing.T) {
	parsed := article.NewParsedFromHTML(
		`<p><a href="https://one.example.com">first</a> and <a href="https://two.example.com">second</a></p>`)
	m := NewWithArticle(parsed, "Title", 72, 100, 30, Options{}, nil)

	require.Len(t, m.links, 2)
	assert.Equal(t, "https://one.example.com", m.links[0].url)
	assert.Equal(t, "https://two.example.com", m.links[1].url)
	assert.Equal(t, "first", spanText(m, 0, 0))
	assert.Equal(t, "second", spanText(m, 1, 0))
}

func TestExtractLinks_SameURLInSeparateParagraphsStaysSeparate(t *testing.T) {
	parsed := article.NewParsedFromHTML(
		`<p>see <a href="https://example.com">here</a></p><p>or <a href="https://example.com">here</a></p>`)
	m := NewWithArticle(parsed, "Title", 72, 100, 30, Options{}, nil)

	assert.Len(t, m.links, 2, "a blank line separates the paragraphs, so no merge")
}

func TestLinkViewable(t *testing.T) {
	assert.True(t, linkViewable("https://example.com/post"))
	assert.True(t, linkViewable("https://example.com/download?file=x.pdf"), "only the path counts, not the query")
	assert.True(t, linkViewable("https://arxiv.org/pdf/2401.12345v2.pdf"), "arXiv PDFs read through the HTML mirror")

	assert.True(t, linkViewable("https://news.ycombinator.com/item?id=42"), "HN discussions open in the comment section")

	assert.False(t, linkViewable("https://example.com/REPORT.PDF"), "case-insensitive")
	assert.False(t, linkViewable("https://example.com/movie.mp4"))
	assert.False(t, linkViewable("https://github.com/schlae/BeavisUltrasound/blob/main/BeavisUltrasoundPnp.pdf"))
	assert.False(t, linkViewable("https://youtube.com/watch?v=1"), "blocked domains can't be rendered either")
	assert.False(t, linkViewable("https://news.ycombinator.com/user?id=dang"), "non-item HN pages have nothing to open")
}

func TestExtractLinks_NoLinks(t *testing.T) {
	m := NewWithArticle(parseTestArticle(t), "Title", 72, 100, 30, Options{}, nil)

	assert.Empty(t, m.links)
}

func TestExtractLinks_HeaderURLRowExcluded(t *testing.T) {
	parsed := article.NewParsedFromHTML(
		`<p>Body with <a href="https://example.com/body">one link</a>.</p>`)

	// The injected header carries its own hyperlink, like the meta block's
	// URL row; it must not become selectable.
	header := func(int) string {
		return "\x1b]8;;https://header.example.com\x1b\\header link\x1b]8;;\x1b\\"
	}

	m := NewWithArticle(parsed, "Title", 72, 100, 30, Options{}, header)

	require.Len(t, m.links, 1)
	assert.Equal(t, "https://example.com/body", m.links[0].url)
}
