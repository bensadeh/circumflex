package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bensadeh/circumflex/cmd"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "gen-completions: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: %s <output-dir>", os.Args[0])
	}

	outDir := filepath.Clean(os.Args[1])
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	root := cmd.Root()
	root.DisableAutoGenTag = true

	if err := root.GenBashCompletionFileV2(filepath.Join(outDir, "clx.bash"), true); err != nil {
		return fmt.Errorf("bash: %w", err)
	}

	if err := root.GenZshCompletionFile(filepath.Join(outDir, "_clx")); err != nil {
		return fmt.Errorf("zsh: %w", err)
	}

	if err := root.GenFishCompletionFile(filepath.Join(outDir, "clx.fish"), true); err != nil {
		return fmt.Errorf("fish: %w", err)
	}

	return nil
}
