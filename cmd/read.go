package cmd

import (
	_ "embed"
	"os"
	"strconv"

	"clx/less"
	"clx/reader"

	"clx/hn/services/hybrid"

	"clx/cli"
	"github.com/spf13/cobra"
)

func readCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "read",
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

			command := cli.Less(article, lesskey.GetPath())

			if err := command.Run(); err != nil {
				defer lesskey.Remove()
				panic(err)
			}
		},
	}
}
