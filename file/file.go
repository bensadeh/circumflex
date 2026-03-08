package file

import (
	"fmt"
	"os"
	"path"
)

const (
	ConfigFileNameFull    = "config.env"
	FavoritesFileNameFull = "favorites.json"
)

func PathToConfigDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}
	configDir := ".config"
	clxDir := "circumflex"

	return path.Join(homeDir, configDir, clxDir)
}

func PathToConfigFile() string {
	return path.Join(PathToConfigDirectory(), ConfigFileNameFull)
}

func PathToFavoritesFile() string {
	return path.Join(PathToConfigDirectory(), FavoritesFileNameFull)
}

func Exists(pathToFile string) bool {
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func ConfigFileExists() bool {
	return Exists(PathToConfigFile())
}

func WriteToFile(path string, content string) error {
	mkdirErr := os.MkdirAll(PathToConfigDirectory(), 0o700)
	if mkdirErr != nil {
		return fmt.Errorf("could not create path to config dir: %w", mkdirErr)
	}

	file, createPathErr := os.Create(path) //nolint:gosec // path from ~/.config/circumflex/
	if createPathErr != nil {
		return fmt.Errorf("could not create config file: %w", createPathErr)
	}
	defer func() { _ = file.Close() }()

	_, writeFileErr := file.WriteString(content)
	if writeFileErr != nil {
		return fmt.Errorf("could not write to file: %w", writeFileErr)
	}

	return nil
}

func WriteToDir(dirPath string, fileName string, content string) error {
	mkdirErr := os.MkdirAll(dirPath, 0o700)
	if mkdirErr != nil {
		return fmt.Errorf("could not create path to config dir: %w", mkdirErr)
	}

	filePath := path.Join(dirPath, fileName)

	file, createPathErr := os.Create(filePath) //nolint:gosec // path from ~/.cache/circumflex/
	if createPathErr != nil {
		return fmt.Errorf("could not create config file: %w", createPathErr)
	}
	defer func() { _ = file.Close() }()

	_, writeFileErr := file.WriteString(content)
	if writeFileErr != nil {
		return fmt.Errorf("could not write to file: %w", writeFileErr)
	}

	return nil
}
