package comment

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/style"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

// renderBody is the body pipeline under golden test. The goldens were
// captured from the legacy string pipeline; the block pipeline must
// reproduce them byte for byte.
func renderBody(html string, commentWidth, screenWidth int, nerdFonts bool, fg color.Color) string {
	return RenderBlocks(Parse(html), RenderOptions{
		CommentWidth: commentWidth,
		ScreenWidth:  screenWidth,
		NerdFonts:    nerdFonts,
		Fg:           fg,
	})
}

// Goldens keep raw ANSI bytes: the rewrite must be byte-identical, styling
// included, not merely layout-identical.
func TestGoldenFixtures(t *testing.T) {
	t.Parallel()

	fixtures := []struct {
		name string
		mod  bool // render with the mod foreground tint
	}{
		{name: "paragraphs"},
		{name: "quotes"},
		{name: "code"},
		{name: "links"},
		{name: "inline"},
		{name: "yc"},
		{name: "deleted"},
		{name: "leadingp"},
		{name: "mod", mod: true},
		{name: "italic-tokens"},
	}

	variants := []struct {
		commentWidth, screenWidth int
		nerd                      bool
	}{
		{commentWidth: 72, screenWidth: 80, nerd: false},
		{commentWidth: 72, screenWidth: 80, nerd: true},
		{commentWidth: 40, screenWidth: 44, nerd: false},
		{commentWidth: 40, screenWidth: 44, nerd: true},
	}

	for _, tt := range fixtures {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src, err := os.ReadFile(filepath.Join("testdata", tt.name+".html"))
			require.NoError(t, err)

			// Fixture files end with one editor-mandated newline that is not
			// part of the comment HTML.
			input := strings.TrimSuffix(string(src), "\n")

			var fg color.Color
			if tt.mod {
				fg = style.CommentModFg()
			}

			var sb strings.Builder

			for _, v := range variants {
				fmt.Fprintf(&sb, "=== width=%d screen=%d nerd=%t ===\n", v.commentWidth, v.screenWidth, v.nerd)
				sb.WriteString(renderBody(input, v.commentWidth, v.screenWidth, v.nerd, fg))
				sb.WriteString("\n")
			}

			compareGolden(t, filepath.Join("testdata", tt.name+".golden"), sb.String())
		})
	}
}

// TestGoldenThread renders every comment body of HN story 48849066 (1079
// comments), the corner-case sweep complementing the targeted fixtures.
//
// Corpus refresh:
//
//	curl -sS "https://hn.algolia.com/api/v1/items/48849066" |
//	  jq -c '[recurse(.children[]) | select(.type=="comment") | {id, author, text}]' |
//	  gzip -n > testdata/thread-48849066.json.gz
func TestGoldenThread(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile(filepath.Join("testdata", "thread-48849066.json.gz"))
	require.NoError(t, err)

	zr, err := gzip.NewReader(bytes.NewReader(src))
	require.NoError(t, err)

	var corpus []struct {
		ID     int    `json:"id"`
		Author string `json:"author"`
		Text   string `json:"text"`
	}

	require.NoError(t, json.NewDecoder(zr).Decode(&corpus))
	require.NotEmpty(t, corpus)

	var sb strings.Builder

	for _, c := range corpus {
		var fg color.Color
		if IsMod(c.Author) {
			fg = style.CommentModFg()
		}

		fmt.Fprintf(&sb, "=== %d %s ===\n", c.ID, c.Author)
		sb.WriteString(renderBody(c.Text, 72, 80, false, fg))
		sb.WriteString("\n")
	}

	compareGolden(t, filepath.Join("testdata", "thread-48849066.golden"), sb.String())
}

func compareGolden(t *testing.T, path, got string) {
	t.Helper()

	if *update {
		require.NoError(t, os.WriteFile(path, []byte(got), 0o600))

		return
	}

	want, err := os.ReadFile(path)
	require.NoError(t, err, "golden file missing, run: go test ./comment/ -update")

	if string(want) == got {
		return
	}

	// Full diffs of ANSI-laden multi-hundred-KB strings are unreadable;
	// report the first divergence with quoted context instead.
	off := 0
	for off < len(want) && off < len(got) && want[off] == got[off] {
		off++
	}

	start := max(0, off-80)
	wantEnd := min(len(want), off+80)
	gotEnd := min(len(got), off+80)

	assert.Fail(t, "golden mismatch",
		"%s: first divergence at byte %d\nwant: %q\ngot:  %q",
		path, off, want[start:wantEnd], got[start:gotEnd])
}
