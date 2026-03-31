package settings

import (
	"os"
	"path/filepath"
)

const (
	favoritesFileNameFull = "favorites.json"

	clxDir = "circumflex"
)

func configDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.TempDir(), clxDir)
	}

	return filepath.Join(dir, clxDir)
}

func CachePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), clxDir)
	}

	return filepath.Join(dir, clxDir)
}

func FavoritesPath() string {
	return filepath.Join(configDir(), favoritesFileNameFull)
}
