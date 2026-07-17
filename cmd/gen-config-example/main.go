package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bensadeh/circumflex/settings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "gen-config-example: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: %s <output-file>", os.Args[0])
	}

	return os.WriteFile(filepath.Clean(os.Args[1]), settings.ExampleConfig(), 0o600)
}
