package file

import (
	"clx/constants/settings"
	"fmt"
	"os"
	"path"
)

func PathToConfigDirectory() string {
	homeDir, _ := os.UserHomeDir()
	configDir := ".config"
	clxDir := "circumflex"

	return path.Join(homeDir, configDir, clxDir)
}

func PathToConfigFile() string {
	return path.Join(PathToConfigDirectory(), settings.ConfigFileNameFull)
}

func PathToFavoritesFile() string {
	return path.Join(PathToConfigDirectory(), settings.FavoritesFileNameFull)
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
	if Exists(path) {
		return nil
	}

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
