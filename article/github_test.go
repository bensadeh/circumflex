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

func TestIsGitHubIssue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url  string
		want bool
	}{
		{url: "https://github.com/adamtwiss/coda/issues/15", want: true},
		{url: "https://github.com/adamtwiss/coda/issues/15/", want: true},
		{url: "https://www.github.com/golang/go/issues/15292", want: true},
		{url: "https://github.com/golang/go/issues/15292#issuecomment-123", want: true},
		{url: "https://github.com/golang/go/pull/15292", want: false},
		{url: "https://github.com/golang/go/issues", want: false},
		{url: "https://github.com/golang/go/issues/new", want: false},
		{url: "https://gist.github.com/user/abc123", want: false},
		{url: "https://example.com/owner/repo/issues/15", want: false},
	}

	for _, tt := range tests {
		u, err := nurl.Parse(tt.url)
		require.NoError(t, err)
		assert.Equal(t, tt.want, isGitHubIssue(u), tt.url)
	}
}

func TestGitHubIssueGolden(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile(filepath.Join("testdata", "github.html"))
	require.NoError(t, err)

	base, err := nurl.Parse("https://github.com/spoonco/spoon-knife/issues/15")
	require.NoError(t, err)

	blocks, title := parseGitHubIssueBlocks(src, base)
	require.NotNil(t, blocks)
	assert.Equal(t, "Give the spoon back", title)

	guessCodeLangs(blocks)

	rendered := ansi.Strip(renderBlocks(blocks, goldenWidth, goldenWidth, showImages)) + "\n"

	goldenPath := filepath.Join("testdata", "github.golden")

	if *update {
		require.NoError(t, os.WriteFile(goldenPath, []byte(rendered), 0o600))

		return
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "golden file missing, run: go test ./article/ -update")

	assert.Equal(t, string(want), rendered)
}

// A GitHub issue page without the embedded payload — a login wall, a
// redesign — parses to nothing, and Parse falls back to readability.
func TestGitHubIssueFallback(t *testing.T) {
	t.Parallel()

	base, err := nurl.Parse("https://github.com/octocat/spoon-knife/issues/15")
	require.NoError(t, err)

	pages := []string{
		"<html><body><p>issues are served fresh</p></body></html>",
		`<html><body><script type="application/json" data-target="react-app.embeddedData">{"payload":{}}</script></body></html>`,
		`<html><body><script type="application/json" data-target="react-app.embeddedData">not json</script></body></html>`,
		`<html><body><script data-target="react-app.embeddedData"></script><p>after</p></body></html>`,
	}

	for _, page := range pages {
		blocks, title := parseGitHubIssueBlocks([]byte(page), base)
		assert.Nil(t, blocks, page)
		assert.Empty(t, title, page)
	}
}
