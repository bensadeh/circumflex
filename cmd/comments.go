package cmd

import (
	"clx/cli"
	"clx/convert"
	"clx/tree"
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func commentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "comments",
		Short: "Go directly to the comment section by ID",
		Long: "Directly enter the comment section for a given item without going through the main " +
			"view first",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Argument must be a valid ID")
				os.Exit(1)
			}

			service := newService()

			comments, err := service.FetchComments(cmd.Context(), id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			config := getConfig()

			screenWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			thread := convert.StoryToThread(comments)
			commentTree := tree.Print(thread, config, screenWidth, time.Now().Unix())

			if err := cli.RunLess(cmd.Context(), commentTree, config); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
}
