package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bensadeh/circumflex/theme"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "gen-theme-example: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: %s <output-file>", os.Args[0])
	}

	content, err := theme.ExampleConfig()
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Clean(os.Args[1]), content, 0o600)
}
