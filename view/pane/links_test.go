package pane

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkViewable(t *testing.T) {
	assert.True(t, LinkViewable("https://example.com/post"))
	assert.True(t, LinkViewable("https://example.com/download?file=x.pdf"), "only the path counts, not the query")
	assert.True(t, LinkViewable("https://arxiv.org/pdf/2401.12345v2.pdf"), "arXiv PDFs read through the HTML mirror")

	assert.True(t, LinkViewable("https://news.ycombinator.com/item?id=42"), "HN discussions open in the comment section")

	assert.False(t, LinkViewable("https://example.com/REPORT.PDF"), "case-insensitive")
	assert.False(t, LinkViewable("https://example.com/movie.mp4"))
	assert.False(t, LinkViewable("https://github.com/schlae/BeavisUltrasound/blob/main/BeavisUltrasoundPnp.pdf"))
	assert.False(t, LinkViewable("https://youtube.com/watch?v=1"), "blocked domains can't be rendered either")
	assert.False(t, LinkViewable("https://news.ycombinator.com/user?id=dang"), "non-item HN pages have nothing to open")
}

func link(url string, lines ...int) Link {
	l := Link{URL: url, Viewable: true}
	for _, line := range lines {
		l.Spans = append(l.Spans, Match{Line: line, StartCell: 0, EndCell: 4})
	}

	return l
}

func TestMoveLink_WrapsAndLeavesViewport(t *testing.T) {
	links := []Link{link("a", 2), link("b", 10), link("c", 40)}

	assert.Equal(t, 2, MoveLink(links, 1, 1, 0))
	assert.Equal(t, 0, MoveLink(links, 2, 1, 0), "wraps forward")
	assert.Equal(t, 2, MoveLink(links, 0, -1, 0), "wraps backward")

	assert.Equal(t, 2, MoveLink(links, -1, 1, 20), "empty selection moves to the first link past the viewport top")
	assert.Equal(t, 1, MoveLink(links, -1, -1, 20), "and backward to the last link above it")
	assert.Equal(t, 0, MoveLink(links, -1, 1, 60), "no link past the top wraps to the first")
}

func TestFirstLinkOnScreen(t *testing.T) {
	links := []Link{link("a", 30), link("b", 45, 46)}

	assert.Equal(t, -1, FirstLinkOnScreen(links, 0, 20))
	assert.Equal(t, 0, FirstLinkOnScreen(links, 25, 20))
	assert.Equal(t, 1, FirstLinkOnScreen(links, 44, 20), "a span on any line inside the window counts")
}
