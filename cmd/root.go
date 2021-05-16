package cmd

import (
	"clx/clx"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clx",
	Short: "It's Hacker News in your terminal",
	Long: `circumflex is a command line tool for browsing Hacker News
in your terminal.`,
	Run: func(cmd *cobra.Command, args []string) {
		clx.Run()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
