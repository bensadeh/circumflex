package settings

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	favoritesFileNameFull = "favorites.json"
	themeFileNameFull     = "theme.toml"

	clxDir = "circumflex"
)

// ConfigDir returns the circumflex config directory. It checks
// $XDG_CONFIG_HOME first, then falls back to os.UserConfigDir().
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, clxDir)
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.TempDir(), clxDir)
	}

	return filepath.Join(dir, clxDir)
}

// CachePath returns the circumflex cache directory. It checks
// $XDG_CACHE_HOME first, then falls back to os.UserCacheDir().
func CachePath() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, clxDir)
	}

	dir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), clxDir)
	}

	return filepath.Join(dir, clxDir)
}

// ThemePath returns the full path to the theme config file.
func ThemePath() string {
	return filepath.Join(ConfigDir(), themeFileNameFull)
}

func FavoritesPath() string {
	return filepath.Join(ConfigDir(), favoritesFileNameFull)
}

// Tilde replaces the user's home directory prefix with ~ for display.
func Tilde(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}

	if path == home {
		return "~"
	}

	if strings.HasPrefix(path, home+string(filepath.Separator)) {
		return "~" + path[len(home):]
	}

	return path
}
