package file

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	favoritesFileNameFull = "favorites.json"

	clxDir         = "circumflex"
	dirPermissions = 0o700
)

func pathToConfigDirectory() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(os.TempDir(), clxDir)
	}

	return filepath.Join(dir, clxDir)
}

func PathToCacheDirectory() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), clxDir)
	}

	return filepath.Join(dir, clxDir)
}

func PathToFavoritesFile() string {
	return filepath.Join(pathToConfigDirectory(), favoritesFileNameFull)
}

func Exists(pathToFile string) bool {
	_, err := os.Stat(pathToFile)

	return err == nil
}

func WriteToFile(filePath string, content string) error {
	dir := filepath.Dir(filePath)

	mkdirErr := os.MkdirAll(dir, dirPermissions)
	if mkdirErr != nil {
		return fmt.Errorf("could not create directory %s: %w", dir, mkdirErr)
	}

	file, createPathErr := os.Create(filePath)
	if createPathErr != nil {
		return fmt.Errorf("could not create file: %w", createPathErr)
	}

	_, writeFileErr := file.WriteString(content)
	if writeFileErr != nil {
		_ = file.Close()

		return fmt.Errorf("could not write to file: %w", writeFileErr)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("could not close file after writing: %w", err)
	}

	return nil
}
