package file

import (
	"fmt"
	"os"
	"path"
)

const (
	FavoritesFileNameFull = "favorites.json"

	clxDir         = "circumflex"
	dirPermissions = 0o700
)

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
	}

	return home
}

func PathToConfigDirectory() string {
	return path.Join(homeDir(), ".config", clxDir)
}

func PathToCacheDirectory() string {
	return path.Join(homeDir(), ".cache", clxDir)
}

func PathToFavoritesFile() string {
	return path.Join(PathToConfigDirectory(), FavoritesFileNameFull)
}

func Exists(pathToFile string) bool {
	_, err := os.Stat(pathToFile)

	return err == nil
}

func WriteToFile(filePath string, content string) error {
	dir := path.Dir(filePath)

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

func WriteToDir(dirPath string, fileName string, content string) error {
	mkdirErr := os.MkdirAll(dirPath, dirPermissions)
	if mkdirErr != nil {
		return fmt.Errorf("could not create path to config dir: %w", mkdirErr)
	}

	filePath := path.Join(dirPath, fileName)

	file, createPathErr := os.Create(filePath)
	if createPathErr != nil {
		return fmt.Errorf("could not create config file: %w", createPathErr)
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
