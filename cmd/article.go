package cmd

import (
	"fmt"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/meta"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/timeago"
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

			style.SetTheme(config.Theme)

			service := newService()

			item, err := service.FetchItem(cmd.Context(), id)
			if err != nil {
				return err
			}

			if item.URL == "" {
				return fmt.Errorf("no link associated with ID %d", id)
			}

			parsed, err := article.Parse(cmd.Context(), item.URL)
			if err != nil {
				return fmt.Errorf("could not read article: %w", err)
			}

			block := meta.ReaderMode(meta.Data{
				URL:       item.URL,
				Author:    item.Author,
				TimeAgo:   timeago.RelativeTime(item.Time),
				ID:        item.ID,
				Points:    item.Points,
				NerdFonts: config.EnableNerdFonts,
			})

			opts := reader.Options{
				URL:       item.URL,
				ID:        item.ID,
				NerdFonts: config.EnableNerdFonts,
				Images:    config.EnableImages,
			}

			return reader.Run(parsed, item.Title, config.ArticleWidth, opts, block.Render)
		},
	}
}
