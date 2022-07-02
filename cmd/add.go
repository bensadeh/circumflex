package cmd

import (
	"strconv"
	"time"

	"clx/bfavorites"
	"clx/item"
	"github.com/charmbracelet/lipgloss"

	"github.com/spf13/cobra"
)

func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add item to list of favorites by ID",
		Long: "Add item to list of favorites by ID. Enter the comment section from inside 'clx' to " +
			"update fields.",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]

			submission := new(item.Item)
			submission.ID, _ = strconv.Atoi(id)
			submission.Title = lipgloss.NewStyle().
				Foreground(lipgloss.Color("3")).
				Render("[Enter comment section to update story]")
			submission.Time = time.Now().Unix()
			submission.User = "[]"

			favorites := bfavorites.New()
			favorites.Add(submission)
			favorites.Write()

			println("Item added to favorites")
			println("(enter the comment section from clx to update)")
		},
	}
}
