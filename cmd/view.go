package cmd

import (
	"clx/cli"
	"clx/comment"
	"clx/config"
	"clx/constants/messages"
	"clx/screen"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(viewCmd)
}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Go directly to the comment section by ID",
	Long:  `Enter the comment section for a given item directly without going through the main view first`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		comments, err := comment.FetchComments(id)
		if err != nil {
			println(messages.CommentsNotFetched)

			os.Exit(1)
		}

		c := config.GetConfig()
		screenWidth := screen.GetTerminalWidth()
		commentTree := comment.ToString(*comments, c.IndentSize, c.CommentWidth, screenWidth, c.PreserveRightMargin,
			c.AltIndentBlock, c.CommentHighlighting)

		cli.Less(commentTree)

		os.Exit(0)
	},
}
