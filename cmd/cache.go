package cmd

import (
	"clx/history"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cacheCmd)
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "cache",
	Run: func(cmd *cobra.Command, args []string) {
		_ = history.Initialize(1)

		println("OK")
	},
}
