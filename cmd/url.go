package cmd

import (
	"fmt"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/view/reader"

	"github.com/spf13/cobra"
)

func urlCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "url [url]",
		Short:                 "open the provided url in reader mode",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := getConfig()
			if err != nil {
				return err
			}

			url := args[0]

			content, err := article.Fetch(cmd.Context(), url, readerWidth(config.ArticleWidth))
			if err != nil {
				return fmt.Errorf("could not read article: %w", err)
			}

			return reader.Run(content, "Reader Mode", reader.Meta{URL: url})
		},
	}
}
