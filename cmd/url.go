package cmd

import (
	_ "embed"
	"os"

	"github.com/bensadeh/circumflex/less"
	"github.com/bensadeh/circumflex/reader"

	"github.com/bensadeh/circumflex/cli"
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

			article, err := reader.GetArticle(url, "Reader Mode", config.CommentWidth, config.IndentationSymbol)
			if err != nil {
				println("Could not read article")
				os.Exit(1)
			}

			lesskey := less.NewLesskey()
			config.LesskeyPath = lesskey.GetPath()

			command := cli.Less(article, config)

			if err := command.Run(); err != nil {
				defer lesskey.Remove()
				panic(err)
			}
		},
	}
}
