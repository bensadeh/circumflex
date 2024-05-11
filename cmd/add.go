package cmd

import (
	"strconv"

	"github.com/f01c33/circumflex/favorites"
	"github.com/f01c33/circumflex/hn/services/hybrid"

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
				panic("ID format error")
			}

			service := hybrid.Service{}
			submission := service.FetchItem(id)

			fav := favorites.New()
			fav.Add(submission)
			fav.Write()

			println("Item added to favorites")
		},
	}
}
