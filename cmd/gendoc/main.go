package main

import (
	"clx/cmd"

	"github.com/spf13/cobra/doc"
)

func main() {
	rootCmd := cmd.Root()

	header := &doc.GenManHeader{
		Title:   "clx",
		Section: "1",
		Source:  "Ben Sadeh",
		Manual:  "circumflex",
	}

	rootCmd.DisableAutoGenTag = true

	if err := doc.GenManTree(rootCmd, header, "./share/man"); err != nil {
		panic(err)
	}
}
