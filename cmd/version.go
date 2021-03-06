package cmd

import (
	"clx/constants/clx"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of circumflex",
	Long:  `Print the version number of circumflex`,
	Run: func(cmd *cobra.Command, args []string) {
		println("circumflex " + clx.Version)
	},
}
