package favorites

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeItem(t *testing.T) {
	item := &Item{
		Title:  "\x1b[31mPWNED",
		Author: "a\x07b",
		URL:    "https://ok.com/\x07path",
		Domain: "ok.com",
	}

	sanitizeItem(item)

	assert.Equal(t, "PWNED", item.Title)
	assert.Equal(t, "ab", item.Author)
	assert.Equal(t, "https://ok.com/path", item.URL)
	assert.Equal(t, "ok.com", item.Domain)
}

// A raw escape byte is rejected by the TOML parser itself, so the reachable
// vector is a six-char unicode escape (assembled from parts so this source
// stays plain text) that TOML decodes into a live ESC for load() to strip.
func TestNew_SanitizesEscapesFromLoadedFavorites(t *testing.T) {
	esc := `\u` + `001b` // decodes to ESC
	bel := `\u` + `0007` // decodes to BEL

	dir := t.TempDir()
	path := filepath.Join(dir, "favorites.toml")

	doc := "[[favorites]]\n" +
		"id = 1\n" +
		`title = "` + esc + `[31mPWNED"` + "\n" +
		`author = "a` + bel + `b"` + "\n" +
		`url = "https://ok.com/` + bel + `path"` + "\n" +
		`domain = "ok.com"` + "\n"
	require.NoError(t, os.WriteFile(path, []byte(doc), 0o600))

	f, err := New(path, "")
	require.NoError(t, err)
	require.Len(t, f.Items(), 1)

	item := f.Items()[0]
	assert.Equal(t, "PWNED", item.Title)
	assert.Equal(t, "ab", item.Author)
	assert.Equal(t, "https://ok.com/path", item.URL)
}

// The pre-5.0 favorites.json may predate sanitization; a unicode escape there
// decodes to a live ESC on migration, which loadLegacyJSON strips.
func TestNew_SanitizesEscapesFromMigratedLegacyJSON(t *testing.T) {
	esc := `\u` + `001b` // decodes to ESC

	dir := t.TempDir()
	legacy := filepath.Join(dir, "favorites.json")
	dest := filepath.Join(dir, "favorites.toml")

	json := `[{"ID":1,"Title":"` + esc + `[31mred","User":"bob","URL":"https://ok.com","Domain":"ok.com"}]`
	require.NoError(t, os.WriteFile(legacy, []byte(json), 0o600))

	f, err := New(dest, legacy)
	require.NoError(t, err)
	require.Len(t, f.Items(), 1)
	assert.Equal(t, "red", f.Items()[0].Title)
}
