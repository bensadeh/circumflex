package article

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/net/html"
)

func normalizedHTML(t *testing.T, src string) string {
	t.Helper()

	node, err := html.Parse(strings.NewReader(src))
	require.NoError(t, err)

	normalizeRoleImages(node)

	var buf bytes.Buffer
	require.NoError(t, html.Render(&buf, node))

	return buf.String()
}

func TestNormalizeRoleImages_InjectsLabelIntoSVGWrapper(t *testing.T) {
	t.Parallel()

	out := normalizedHTML(t, `<div role="img" aria-label="A chart trending upward."><svg><circle></circle></svg></div>`)

	assert.Contains(t, out, "<p>A chart trending upward.</p></div>")
}

func TestNormalizeRoleImages_ReplacesBareSVG(t *testing.T) {
	t.Parallel()

	out := normalizedHTML(t, `<svg role="img" aria-label="A chart trending upward."><circle></circle></svg>`)

	assert.NotContains(t, out, "<svg")
	assert.Contains(t, out, `<div role="img" aria-label="A chart trending upward."><p>A chart trending upward.</p></div>`)
}

// The label is an attribute of an attacker-controlled page that renders as a
// caption, so it is sanitized like the alt text it stands in for.
func TestRoleImageLabel_Sanitized(t *testing.T) {
	t.Parallel()

	blocks := blocksFromHTML(t, `<div role="img" aria-label="chart \x1b[31mred\x1b]0;pwn\x07"><svg></svg></div>`)

	require.Len(t, blocks, 1)
	assert.Equal(t, `chart \x1b[31mred\x1b]0;pwn\x07`, blocks[0].plainText())

	raw := blocksFromHTML(t, "<div role=\"img\" aria-label=\"chart \x1b[31mred\x1b]0;pwn\x07\"><svg></svg></div>")

	require.Len(t, raw, 1)
	assert.Equal(t, "chart red", raw[0].plainText())
}

func TestNormalizeRoleImages_LeavesEmojiSpansAlone(t *testing.T) {
	t.Parallel()

	src := `<p>done <span role="img" aria-label="tada">🎉</span> indeed</p>`
	out := normalizedHTML(t, src)

	assert.Contains(t, out, src)
}
