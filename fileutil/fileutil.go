package fileutil

import (
	"os"
	"path/filepath"
)

// Exists reports whether path exists on disk.
func Exists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

// WriteAtomic writes content to path atomically by writing to a
// temporary file in the same directory and renaming it into place.
// Parent directories are created with mode 0o700 if they don't exist.
func WriteAtomic(path string, content string) error {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp*")
	if err != nil {
		return err
	}

	tmpPath := tmp.Name()

	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)

		return err
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)

		return err
	}

	return os.Rename(tmpPath, path)
}
