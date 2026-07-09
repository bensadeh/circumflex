package cmd

import (
	"fmt"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/meta"
	"github.com/bensadeh/circumflex/style"
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

			style.SetTheme(config.Theme)

			url := args[0]

			parsed, err := article.Parse(cmd.Context(), url)
			if err != nil {
				return fmt.Errorf("could not read article: %w", err)
			}

			articleMeta := reader.Meta{URL: url, NerdFonts: config.EnableNerdFonts, Images: config.EnableImages}

			return reader.Run(parsed, "Reader Mode", config.ArticleWidth, articleMeta,
				func(contentWidth int) string {
					return meta.ReaderModeURLBlock(url, config.EnableNerdFonts, contentWidth)
				})
		},
	}
}
