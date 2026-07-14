package article

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/ansi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/net/html"
)

var update = flag.Bool("update", false, "update golden files")

const goldenWidth = 72

// Goldens are stored ANSI-stripped: layout, wrapping and glyphs are covered
// here, styling is covered by the renderer unit tests.
func TestGolden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		fixture  string
		hostname string
	}{
		{fixture: "article", hostname: "example.com"},
		{fixture: "wikipedia", hostname: "en.wikipedia.org"},
		{fixture: "mediawiki", hostname: "en.wikipedia.org"},
		{fixture: "math", hostname: "blog.example.com"},
		{fixture: "arxiv", hostname: "arxiv.org"},
	}

	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			t.Parallel()

			src, err := os.ReadFile(filepath.Join("testdata", tt.fixture+".html"))
			require.NoError(t, err)

			node, err := html.Parse(strings.NewReader(string(src)))
			require.NoError(t, err)

			normalizeMediaWiki(node)

			blocks := parseBlocks(node)

			if usesMathRenderer(src) {
				convertMath(blocks)
			}

			blocks = applySiteRules(blocks, tt.hostname)

			rendered := ansi.Strip(renderBlocks(blocks, goldenWidth, goldenWidth, showImages)) + "\n"

			goldenPath := filepath.Join("testdata", tt.fixture+".golden")

			if *update {
				require.NoError(t, os.WriteFile(goldenPath, []byte(rendered), 0o600))

				return
			}

			want, err := os.ReadFile(goldenPath)
			require.NoError(t, err, "golden file missing, run: go test ./article/ -update")

			assert.Equal(t, string(want), rendered)
		})
	}
}
