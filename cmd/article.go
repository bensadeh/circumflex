package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/view/reader"

	"github.com/spf13/cobra"
)

func articleCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "article",
		Short:                 "Read the linked article associated with an item based on the ID",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			id, convErr := strconv.Atoi(args[0])
			if convErr != nil {
				fmt.Fprintln(os.Stderr, "Argument must be a valid ID")
				os.Exit(1)
			}

			service := newService()

			item, err := service.FetchItem(cmd.Context(), id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if item.URL == "" {
				fmt.Fprintln(os.Stderr, "Could not find any links associated with the ID "+args[0])
				os.Exit(1)
			}

			config := getConfig()

			content, err := article.Fetch(cmd.Context(), item.URL, readerWidth(config.ArticleWidth))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading article: %v\n", err)
				os.Exit(1)
			}

			if err := reader.Run(content, item.Title); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}
