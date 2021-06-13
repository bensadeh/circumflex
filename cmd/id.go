package cmd

import (
	"clx/cli"
	"clx/comment"
	"clx/constants/messages"
	"clx/screen"
	"clx/settings"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(idCmd)
}

var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Go directly to comment section by ID",
	Long:  `Enter the comment section for a given item directly without going through the main view`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		comments, err := comment.FetchComments(id)
		if err != nil {
			println(messages.CommentsNotFetched)

			os.Exit(1)
		}

		screenWidth := screen.GetTerminalWidth()
		commentTree := comment.ToString(*comments, settings.IndentSizeDefault, settings.CommentWidthDefault,
			screenWidth, settings.PreserveRightMarginDefault, settings.UseAlternateIndentBlockDefault)

		cli.Less(commentTree)

		os.Exit(0)
	},
}
