package cmd

import (
	"strconv"
	"time"

	"clx/cli"
	"clx/comment"
	hybrid_bubble "clx/hn/services/hybrid"
	"clx/screen"
	"clx/settings"

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

			service := new(hybrid_bubble.Service)

			comments := service.FetchStory(id)

			config := settings.New()

			screenWidth := screen.GetTerminalWidth()
			commentTree := comment.ToString(comments, config, screenWidth, time.Now().Unix())

			cli.Less(commentTree)
		},
	}
}
