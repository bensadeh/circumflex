package cmd

import (
	"time"

	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/view/comments"

	"github.com/spf13/cobra"
)

func commentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "comments [id]",
		Short:                 "read the comment section of a story",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}

			config, err := getConfig()
			if err != nil {
				return err
			}

			style.SetTheme(config.Theme)

			service := newService()

			tree, err := service.FetchComments(cmd.Context(), id, nil)
			if err != nil {
				return err
			}

			return comments.Run(comment.ToThread(tree), time.Now().Unix(),
				config.CommentWidth, config.Indent, config.EnableNerdFonts)
		},
	}
}
