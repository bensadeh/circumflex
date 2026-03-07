package cmd

import (
	"clx/history"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func clearCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "clear",
		Short:                 "Clear the history of visited IDs",
		Long:                  "Clear the history of visited IDs from ~/.cache/circumflex/history.json.",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			his := history.Persistent{}
			if err := his.ClearAndWriteToDisk(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			println("List of visited IDs cleared")
		},
	}
}
