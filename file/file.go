package file

import (
	"clx/constants/settings"
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
	} else {
		return true
	}
}

func ConfigFileExists() bool {
	return Exists(PathToConfigFile())
}

func WriteToConfigFile(content string) {
	if Exists(PathToConfigFile()) {
		return
	}

	_ = os.MkdirAll(PathToConfigDirectory(), 0700)

	f, createFileErr := os.Create(PathToConfigFile())
	if createFileErr != nil {
		panic(createFileErr)
	}

	_, writeFileErr := f.WriteString(content)
	if writeFileErr != nil {
		panic(writeFileErr)
	}
}