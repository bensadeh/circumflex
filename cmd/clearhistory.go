package cmd

import (
	"fmt"

	"github.com/bensadeh/circumflex/history"

	"github.com/spf13/cobra"
)

func clearHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "clear-history",
		Aliases:               []string{"clear"},
		Short:                 "clear the history of visited IDs",
		Args:                  cobra.NoArgs,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := history.ClearPersistent(); err != nil {
				return err
			}

			fmt.Println("List of visited IDs cleared")

			return nil
		},
	}
}
