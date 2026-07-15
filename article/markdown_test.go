package article

import (
	nurl "net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/bensadeh/circumflex/ansi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownGolden(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile(filepath.Join("testdata", "markdown.md"))
	require.NoError(t, err)

	base, err := nurl.Parse("https://example.com/posts/structured-text.md")
	require.NoError(t, err)

	blocks, title, err := parseMarkdownBlocks(src, base)
	require.NoError(t, err)

	assert.Equal(t, "Structured Text Rendering", title)

	rendered := ansi.Strip(renderBlocks(blocks, goldenWidth, goldenWidth, showImages)) + "\n"

	goldenPath := filepath.Join("testdata", "markdown.golden")

	if *update {
		require.NoError(t, os.WriteFile(goldenPath, []byte(rendered), 0o600))

		return
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "golden file missing, run: go test ./article/ -update")

	assert.Equal(t, string(want), rendered)
}

func TestIsMarkdown(t *testing.T) {
	t.Parallel()

	mdURL, err := nurl.Parse("https://example.com/post.md")
	require.NoError(t, err)

	htmlURL, err := nurl.Parse("https://example.com/post")
	require.NoError(t, err)

	tests := []struct {
		name        string
		contentType string
		url         *nurl.URL
		body        string
		want        bool
	}{
		{"text/markdown header", "text/markdown; charset=utf-8", htmlURL, "# Hi", true},
		{"text/plain with .md path", "text/plain; charset=utf-8", mdURL, "# Hi", true},
		{"text/plain without .md path", "text/plain; charset=utf-8", htmlURL, "# Hi", false},
		{"mislabeled html with .md path", "text/plain", mdURL, "<!DOCTYPE html><html>", false},
		{"html", "text/html; charset=utf-8", mdURL, "<!DOCTYPE html>", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, isMarkdown(tt.contentType, tt.url, []byte(tt.body)))
		})
	}
}

func TestMarkdownResolvesRelativeLinks(t *testing.T) {
	t.Parallel()

	base, err := nurl.Parse("https://example.com/posts/entry.md")
	require.NoError(t, err)

	blocks, _, err := parseMarkdownBlocks([]byte("[relative](../other/) and [anchored](#note)"), base)
	require.NoError(t, err)

	require.Len(t, blocks, 1)

	var hrefs []string

	for _, s := range blocks[0].spans {
		if s.href != "" {
			hrefs = append(hrefs, s.href)
		}
	}

	assert.Equal(t, []string{"https://example.com/other/"}, hrefs)
}
