package cmd

import (
	"clx/constants/clx"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:                   "version",
	Short:                 "Print the version number of circumflex",
	Long:                  `Print the version number of circumflex`,
	Args:                  cobra.NoArgs,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		println("circumflex " + clx.Version)
	},
}
