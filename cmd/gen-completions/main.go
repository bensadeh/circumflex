package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bensadeh/circumflex/cmd"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <output-dir>\n", os.Args[0])
		os.Exit(1)
	}

	outDir := filepath.Clean(os.Args[1])
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "could not create %s: %v\n", outDir, err)
		os.Exit(1)
	}

	root := cmd.Root()
	root.DisableAutoGenTag = true

	if err := root.GenBashCompletionFileV2(filepath.Join(outDir, "clx.bash"), true); err != nil {
		fmt.Fprintf(os.Stderr, "bash: %v\n", err)
		os.Exit(1)
	}

	if err := root.GenZshCompletionFile(filepath.Join(outDir, "_clx")); err != nil {
		fmt.Fprintf(os.Stderr, "zsh: %v\n", err)
		os.Exit(1)
	}

	if err := root.GenFishCompletionFile(filepath.Join(outDir, "clx.fish"), true); err != nil {
		fmt.Fprintf(os.Stderr, "fish: %v\n", err)
		os.Exit(1)
	}
}
