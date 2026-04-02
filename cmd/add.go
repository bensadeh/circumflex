package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/settings"

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

			service := newService()

			submission, err := service.FetchItem(cmd.Context(), id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fav := favorites.New(settings.FavoritesPath())
			fav.Add(submission)

			if err := fav.Write(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Item added to favorites")
		},
	}
}
