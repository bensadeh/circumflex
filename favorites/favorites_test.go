package favorites

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFavorites_AddAndItems(t *testing.T) {
	f := &Favorites{}
	assert.Empty(t, f.Items())

	f.Add(&Item{ID: 1, Title: "First"})
	assert.Len(t, f.Items(), 1)

	f.Add(&Item{ID: 2, Title: "Second"})
	assert.Len(t, f.Items(), 2)
}

func TestFavorites_Remove(t *testing.T) {
	f := &Favorites{}
	f.Add(&Item{ID: 1, Title: "First"})
	f.Add(&Item{ID: 2, Title: "Second"})
	f.Add(&Item{ID: 3, Title: "Third"})

	err := f.Remove(1)
	require.NoError(t, err)
	assert.Len(t, f.Items(), 2)
	assert.Equal(t, 1, f.Items()[0].ID)
	assert.Equal(t, 3, f.Items()[1].ID)
}

func TestFavorites_Remove_First(t *testing.T) {
	f := &Favorites{}
	f.Add(&Item{ID: 1, Title: "First"})
	f.Add(&Item{ID: 2, Title: "Second"})

	err := f.Remove(0)
	require.NoError(t, err)
	assert.Len(t, f.Items(), 1)
	assert.Equal(t, 2, f.Items()[0].ID)
}

func TestFavorites_Remove_Last(t *testing.T) {
	f := &Favorites{}
	f.Add(&Item{ID: 1, Title: "First"})
	f.Add(&Item{ID: 2, Title: "Second"})

	err := f.Remove(1)
	require.NoError(t, err)
	assert.Len(t, f.Items(), 1)
	assert.Equal(t, 1, f.Items()[0].ID)
}

func TestFavorites_Remove_OutOfBounds(t *testing.T) {
	f := &Favorites{}
	f.Add(&Item{ID: 1, Title: "First"})

	err := f.Remove(-1)
	require.Error(t, err)

	err = f.Remove(1)
	require.Error(t, err)

	err = f.Remove(5)
	require.Error(t, err)
}

func TestFavorites_Remove_Empty(t *testing.T) {
	f := &Favorites{}

	err := f.Remove(0)
	assert.Error(t, err)
}

func TestLoadLegacyJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	require.NoError(t, os.WriteFile(path,
		[]byte(`[{"ID":1,"Title":"Test","Points":100,"User":"dang"}]`), 0o600))

	items, err := loadLegacyJSON(path)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, 1, items[0].ID)
	assert.Equal(t, "Test", items[0].Title)
	assert.Equal(t, "dang", items[0].Author, `legacy "User" key maps to Author`)
}

func TestLoadLegacyJSON_Invalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "favorites.json")
	require.NoError(t, os.WriteFile(path, []byte(`not json`), 0o600))

	items, err := loadLegacyJSON(path)
	require.Error(t, err)
	assert.Nil(t, items)
}

func TestFavorites_UpdateStoryAndWriteToDisk(t *testing.T) {
	f := &Favorites{path: filepath.Join(t.TempDir(), "favorites.toml")}
	f.Add(&Item{ID: 1, Title: "Old Title", Points: 10, CommentsCount: 5})

	newItem := &Item{ID: 1, Title: "New Title", Points: 100, CommentsCount: 50}
	require.NoError(t, f.UpdateStoryAndWriteToDisk(newItem))

	assert.Equal(t, "New Title", f.Items()[0].Title)
	assert.Equal(t, 100, f.Items()[0].Points)
	assert.Equal(t, 50, f.Items()[0].CommentsCount)
}

func TestFavorites_UpdateStoryAndWriteToDisk_NoMatch(t *testing.T) {
	f := &Favorites{path: filepath.Join(t.TempDir(), "favorites.toml")}
	f.Add(&Item{ID: 1, Title: "Original"})

	newItem := &Item{ID: 99, Title: "Different"}
	require.NoError(t, f.UpdateStoryAndWriteToDisk(newItem))

	assert.Equal(t, "Original", f.Items()[0].Title)
}

func favoritesPaths(t *testing.T) (string, string) {
	t.Helper()

	dir := t.TempDir()

	return filepath.Join(dir, "favorites.toml"), filepath.Join(dir, "favorites.json")
}

func TestNew_CorruptFile(t *testing.T) {
	path, legacy := favoritesPaths(t)
	require.NoError(t, os.WriteFile(path, []byte("not valid toml{{{"), 0o600))

	f, err := New(path, legacy)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "corrupted")
	assert.NotNil(t, f)
	assert.Empty(t, f.Items())
}

func TestNew_UnknownKeyFails(t *testing.T) {
	path, legacy := favoritesPaths(t)
	require.NoError(t, os.WriteFile(path,
		[]byte("[[favorites]]\nid = 1\ntitle = \"Saved\"\ncomment_count = 3\n"), 0o600))

	_, err := New(path, legacy)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "comment_count")
}

func TestNew_ValidFile(t *testing.T) {
	path, legacy := favoritesPaths(t)
	require.NoError(t, os.WriteFile(path,
		[]byte("[[favorites]]\nid = 1\ntitle = \"Saved\"\nauthor = \"dang\"\n"), 0o600))

	f, err := New(path, legacy)

	require.NoError(t, err)
	assert.Len(t, f.Items(), 1)
	assert.Equal(t, "Saved", f.Items()[0].Title)
	assert.Equal(t, "dang", f.Items()[0].Author)
}

func TestNew_MissingFile(t *testing.T) {
	path, legacy := favoritesPaths(t)

	f, err := New(path, legacy)

	require.NoError(t, err)
	assert.Empty(t, f.Items())
	assert.NoFileExists(t, path, "nothing to migrate, nothing written")
}

func TestNew_MigratesLegacyJSON(t *testing.T) {
	path, legacy := favoritesPaths(t)
	require.NoError(t, os.WriteFile(legacy,
		[]byte(`[{"ID":1,"Title":"Old","User":"dang","CommentsCount":7}]`), 0o600))

	f, err := New(path, legacy)

	require.NoError(t, err)
	assert.Len(t, f.Items(), 1)
	assert.Equal(t, "dang", f.Items()[0].Author)
	assert.FileExists(t, path, "migration writes favorites.toml")
	assert.FileExists(t, legacy, "the JSON original is left untouched")

	reloaded, err := New(path, legacy)
	require.NoError(t, err)
	assert.Equal(t, f.Items(), reloaded.Items())
}

func TestNew_ExistingTomlWinsOverLegacy(t *testing.T) {
	path, legacy := favoritesPaths(t)
	require.NoError(t, os.WriteFile(path, []byte("[[favorites]]\nid = 2\ntitle = \"New\"\n"), 0o600))
	require.NoError(t, os.WriteFile(legacy, []byte(`[{"ID":1,"Title":"Old"}]`), 0o600))

	f, err := New(path, legacy)

	require.NoError(t, err)
	assert.Len(t, f.Items(), 1)
	assert.Equal(t, "New", f.Items()[0].Title)
}

func TestAtomicWrite_RoundTrip(t *testing.T) {
	path, legacy := favoritesPaths(t)

	f := &Favorites{path: path}
	f.Add(&Item{ID: 1, Title: "Test", Author: "dang", CommentsCount: 3})

	require.NoError(t, f.Write())

	f2, err := New(path, legacy)
	require.NoError(t, err)
	assert.Len(t, f2.Items(), 1)
	assert.Equal(t, f.Items(), f2.Items())
}
