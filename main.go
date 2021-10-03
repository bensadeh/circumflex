package main

import (
	"clx/cmd"
)

func main() {
	rootCmd := cmd.Root()
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
