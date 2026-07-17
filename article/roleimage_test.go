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

func TestNormalizeRoleImages_LeavesEmojiSpansAlone(t *testing.T) {
	t.Parallel()

	src := `<p>done <span role="img" aria-label="tada">🎉</span> indeed</p>`
	out := normalizedHTML(t, src)

	assert.Contains(t, out, src)
}
