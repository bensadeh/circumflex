package cmd

import (
	"fmt"
	"os"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/view/reader"

	"github.com/spf13/cobra"
)

func urlCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "url [url]",
		Short:                 "Open the provided url in reader mode",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			config := getConfig()
			url := args[0]

			content, err := article.Fetch(cmd.Context(), url, readerWidth(config.ArticleWidth))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading article: %v\n", err)
				os.Exit(1)
			}

			if err := reader.Run(content, "Reader Mode"); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}
