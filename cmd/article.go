package cmd

import (
	_ "embed"
	"os"
	"strconv"

	"github.com/f01c33/circumflex/less"
	"github.com/f01c33/circumflex/reader"

	"github.com/f01c33/circumflex/hn/services/hybrid"

	"github.com/f01c33/circumflex/cli"
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
				println("Argument must be a valid ID")
				os.Exit(1)
			}

			service := new(hybrid.Service)

			item := service.FetchItem(id)

			if item.URL == "" {
				println("Could not find any links associated with the ID " + args[0])
				os.Exit(1)
			}

			config := getConfig()

			article, _ := reader.GetArticle(item.URL, item.Title, config.CommentWidth, config.IndentationSymbol)

			lesskey := less.NewLesskey()

			command := cli.Less(article, config)

			if err := command.Run(); err != nil {
				defer lesskey.Remove()
				panic(err)
			}
		},
	}
}
