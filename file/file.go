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

func Exists(pathToFile string) bool {
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func ConfigFileExists() bool {
	return Exists(PathToConfigFile())
}

func WriteToConfigFile(content string) error {
	if Exists(PathToConfigFile()) {
		return nil
	}

	mkdirErr := os.MkdirAll(PathToConfigDirectory(), 0o700)
	if mkdirErr != nil {
		return fmt.Errorf("could not create path to config dir: %w", mkdirErr)
	}

	f, createFileErr := os.Create(PathToConfigFile())
	if createFileErr != nil {
		return fmt.Errorf("could not create config file: %w", createFileErr)
	}

	_, writeFileErr := f.WriteString(content)
	if writeFileErr != nil {
		return fmt.Errorf("could not write to file: %w", writeFileErr)
	}

	return nil
}
