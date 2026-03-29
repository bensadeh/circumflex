package cmd

import (
	readerView "clx/bubble/reader"
	"clx/reader"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func urlCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "url",
		Short:                 "Open the provided url in reader mode in the terminal",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			config := getConfig()
			url := args[0]

			article, err := reader.Article(cmd.Context(), url, "Reader Mode", config.CommentWidth, config.IndentationSymbol)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Could not read article")
				os.Exit(1)
			}

			if err := readerView.Run(article, "Reader Mode"); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}
