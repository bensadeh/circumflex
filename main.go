package main

import "github.com/bensadeh/circumflex/cmd"

func main() {
	rootCmd := cmd.Root()
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
