package cmd

import (
	"clx/cli"
	"clx/less"
	"clx/reader"
	_ "embed"
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

			article, err := reader.GetArticle(cmd.Context(), url, "Reader Mode", config.CommentWidth, config.IndentationSymbol)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Could not read article")
				os.Exit(1)
			}

			lesskey, err := less.NewLesskey()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not create lesskey: %v\n", err)
				os.Exit(1)
			}
			defer lesskey.Remove()

			config.LesskeyPath = lesskey.GetPath()

			command := cli.Less(cmd.Context(), article, config)

			if err := command.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Could not run less: %v\n", err)
				os.Exit(1)
			}
		},
	}
}
