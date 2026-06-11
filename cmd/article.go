package cmd

import (
	"fmt"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/view/reader"

	"github.com/spf13/cobra"
)

func articleCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "article [id]",
		Short:                 "read the linked article of a story",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}

			config, err := getConfig()
			if err != nil {
				return err
			}

			service := newService()

			item, err := service.FetchItem(cmd.Context(), id)
			if err != nil {
				return err
			}

			if item.URL == "" {
				return fmt.Errorf("no link associated with ID %d", id)
			}

			content, err := article.Fetch(cmd.Context(), item.URL, readerWidth(config.ArticleWidth))
			if err != nil {
				return fmt.Errorf("could not read article: %w", err)
			}

			return reader.Run(content, item.Title, reader.Meta{URL: item.URL, ID: item.ID})
		},
	}
}
