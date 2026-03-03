package cmd

import (
	"fmt"
	"os"
	"strconv"

	"clx/favorites"
	"clx/hn/services/hybrid"

	"github.com/spf13/cobra"
)

func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "add",
		Short:                 "Add item to list of favorites by ID",
		Long:                  "Add item to list of favorites by ID",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Argument must be a valid ID")
				os.Exit(1)
			}

			service := hybrid.Service{}
			submission, err := service.FetchItem(id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fav := favorites.New()
			fav.Add(submission)
			fav.Write()

			println("Item added to favorites")
		},
	}
}
