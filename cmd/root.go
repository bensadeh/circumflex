package cmd

import (
	"clx/bubble"
	clx2 "clx/constants/clx"
	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "clx",
		Short:   "circumflex is a command line tool for browsing Hacker News in your terminal",
		Long:    "circumflex is a command line tool for browsing Hacker News in your terminal",
		Version: clx2.Version,
		Run: func(cmd *cobra.Command, args []string) {
			bubble.Run()
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(clearCmd())
	rootCmd.AddCommand(viewCmd())
	rootCmd.AddCommand(legacyCmd())

	return rootCmd
}
