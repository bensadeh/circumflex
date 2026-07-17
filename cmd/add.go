package cmd

import (
	"fmt"

	"github.com/bensadeh/circumflex/favorites"
	"github.com/bensadeh/circumflex/settings"

	"github.com/spf13/cobra"
)

func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "add [id]",
		Short:                 "add item to list of favorites",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}

			service := newService()

			story, err := service.FetchItem(cmd.Context(), id)
			if err != nil {
				return err
			}

			fav, err := favorites.New(settings.FavoritesPath(), settings.LegacyFavoritesPath())
			if err != nil {
				return err
			}

			fav.Add(favorites.ItemFromStory(story))

			if err := fav.Write(); err != nil {
				return err
			}

			fmt.Println("Item added to favorites")

			return nil
		},
	}
}
