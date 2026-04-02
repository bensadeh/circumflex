package main

import (
	"os"

	"github.com/bensadeh/circumflex/cmd"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		os.Exit(1)
	}
}
