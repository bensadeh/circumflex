package main

import (
	"clx/cmd"
	"os"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		os.Exit(1)
	}
}
