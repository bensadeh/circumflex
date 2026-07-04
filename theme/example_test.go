package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExampleConfigMatchesRepoFile(t *testing.T) {
	t.Parallel()

	want, err := ExampleConfig()
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join("..", "theme.toml.example"))
	require.NoError(t, err)

	require.Equal(t, string(want), string(got),
		"theme.toml.example is out of date — run share/generate_all.sh")
}
