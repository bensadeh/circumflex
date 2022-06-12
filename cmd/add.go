package cmd

import (
	"clx/constants/messages"
	"clx/favorites"
	"clx/handler"
	"clx/history"
	"clx/item"
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"time"

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
			fav := favorites.Initialize()
			his := history.NewNonPersistentHistory()
			sh := new(handler.StoryHandler)
			sh.Init(fav, his)

			submission := new(item.Item)
			submission.ID, _ = strconv.Atoi(id)
			submission.Title = lipgloss.NewStyle().
				Foreground(lipgloss.Color("3")).
				Render(messages.EnterCommentSectionToUpdate)
			submission.Time = time.Now().Unix()
			submission.User = "[]"

			_ = sh.AddItemToFavoritesAndWriteToFile(submission)

			println("Item added to favorites")
			println("(enter the comment section from clx to update)")
		},
	}
}
