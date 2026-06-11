package fileutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAtomic_CreatesFileAndParentDirs(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nested", "dir", "file.json")

	require.NoError(t, WriteAtomic(path, "content"))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "content", string(got))
}

func TestWriteAtomic_OverwritesExistingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "file.json")

	require.NoError(t, WriteAtomic(path, "old"))
	require.NoError(t, WriteAtomic(path, "new"))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "new", string(got))
}

func TestWriteAtomic_LeavesNoTempFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	require.NoError(t, WriteAtomic(filepath.Join(dir, "file.json"), "content"))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, e := range entries {
		assert.NotContains(t, e.Name(), ".tmp", "temp file left behind: %s", e.Name())
	}
}

func TestWriteAtomic_FailedWriteKeepsOriginal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "file.json")

	require.NoError(t, WriteAtomic(path, "original"))

	// A path whose parent is a regular file cannot be created.
	require.Error(t, WriteAtomic(filepath.Join(path, "child.json"), "x"))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "original", string(got))
}

func TestExists(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "file.json")
	assert.False(t, Exists(path))

	require.NoError(t, os.WriteFile(path, []byte("x"), 0o600))
	assert.True(t, Exists(path))
}
