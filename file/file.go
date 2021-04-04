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
	homeDir, _ := os.UserHomeDir()
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

	file, createFileErr := os.Create(path)
	if createFileErr != nil {
		return fmt.Errorf("could not create config file: %w", createFileErr)
	}

	_, writeFileErr := file.WriteString(content)
	if writeFileErr != nil {
		return fmt.Errorf("could not write to file: %w", writeFileErr)
	}

	return nil
}
