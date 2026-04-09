package settings

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigDir_XDGOverride(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	assert.Equal(t, filepath.Join(xdg, "circumflex"), ConfigDir())
}

func TestConfigDir_FallbackWhenXDGUnset(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	dir := ConfigDir()

	// Should fall back to the platform default (os.UserConfigDir).
	platformDir, err := os.UserConfigDir()
	if err != nil {
		// If os.UserConfigDir fails, we fall back to TempDir.
		assert.Equal(t, filepath.Join(os.TempDir(), "circumflex"), dir)

		return
	}

	assert.Equal(t, filepath.Join(platformDir, "circumflex"), dir)
}

func TestCachePath_XDGOverride(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", xdg)

	assert.Equal(t, filepath.Join(xdg, "circumflex"), CachePath())
}

func TestCachePath_FallbackWhenXDGUnset(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "")

	dir := CachePath()

	platformDir, err := os.UserCacheDir()
	if err != nil {
		assert.Equal(t, filepath.Join(os.TempDir(), "circumflex"), dir)

		return
	}

	assert.Equal(t, filepath.Join(platformDir, "circumflex"), dir)
}

func TestThemePath_UsesConfigDir(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	assert.Equal(t, filepath.Join(xdg, "circumflex", "theme.toml"), ThemePath())
}

func TestFavoritesPath_UsesConfigDir(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	assert.Equal(t, filepath.Join(xdg, "circumflex", "favorites.json"), FavoritesPath())
}

func TestConfigDir_FallbackOnMacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-only test")
	}

	t.Setenv("XDG_CONFIG_HOME", "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	expected := filepath.Join(home, "Library", "Application Support", "circumflex")
	assert.Equal(t, expected, ConfigDir())
}

func TestCachePath_FallbackOnMacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-only test")
	}

	t.Setenv("XDG_CACHE_HOME", "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	expected := filepath.Join(home, "Library", "Caches", "circumflex")
	assert.Equal(t, expected, CachePath())
}
