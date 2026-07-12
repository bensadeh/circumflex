package cmd

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

// The help output is assembled from pflag data with hand-rolled alignment,
// coloring, wrapping and default rendering; the golden pins all of it.
func TestHelp_MatchesGolden(t *testing.T) {
	t.Setenv("CLX_READER_MODE_IMAGES", "")

	var buf bytes.Buffer

	root := Root()
	root.SetOut(&buf)
	require.NoError(t, root.Help())

	golden := filepath.Join("testdata", "help.golden")

	if *update {
		require.NoError(t, os.MkdirAll("testdata", 0o750))
		require.NoError(t, os.WriteFile(golden, buf.Bytes(), 0o600))
	}

	want, err := os.ReadFile(golden)
	require.NoError(t, err, "golden file missing, run: go test ./cmd/ -update")
	assert.Equal(t, string(want), buf.String(), "help output changed — run: go test ./cmd/ -update")
}
