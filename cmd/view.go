package cmd

import (
	"clx/cli"
	"clx/comment"
	"clx/hn/services/hybrid"
	"clx/screen"
	"clx/settings"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func viewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Go directly to the comment section by ID",
		Long: "Directly enter the comment section for a given item without going through the main " +
			"view first",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := strconv.Atoi(args[0])

			service := new(hybrid.Service)

			comments := service.FetchStory(id)
			//if err != nil {
			//	println(messages.CommentsNotFetched)
			//
			//	os.Exit(1)
			//}

			config := settings.New()

			screenWidth := screen.GetTerminalWidth()
			commentTree := comment.ToString(comments, config, screenWidth, time.Now().Unix())

			cli.Less(commentTree)
		},
	}
}
