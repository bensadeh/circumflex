package cmd

//import (
//	"clx/cli"
//	"clx/comment"
//	"clx/constants/messages"
//	"clx/core"
//	"clx/screen"
//	"os"
//
//	"github.com/spf13/cobra"
//)
//
//func viewCmd() *cobra.Command {
//	return &cobra.Command{
//		Use:   "view",
//		Short: "Go directly to the comment section by ID",
//		Long: "Directly enter the comment section for a given item without going through the main " +
//			"view first",
//		Args:                  cobra.ExactArgs(1),
//		DisableFlagsInUseLine: true,
//		Run: func(cmd *cobra.Command, args []string) {
//			id := args[0]
//
//			comments, err := comment.FetchComments(id)
//			if err != nil {
//				println(messages.CommentsNotFetched)
//
//				os.Exit(1)
//			}
//
//			config := core.GetConfigWithDefaults()
//
//			screenWidth := screen.GetTerminalWidth()
//			commentTree := comment.ToString(*comments, config, screenWidth)
//
//			cli.Less(commentTree)
//		},
//	}
//}
