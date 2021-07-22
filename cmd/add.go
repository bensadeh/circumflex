package cmd

import (
	"clx/constants/messages"
	"clx/endpoints"
	"clx/favorites"
	"clx/handler"
	"clx/history"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add item to list of favorites by ID",
	Long:  "Add item to list of favorites by ID. Enter the comment section from inside 'clx' to update fields.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		fav := favorites.Initialize()
		his := history.Initialize(0)
		sh := new(handler.StoryHandler)
		sh.Init(fav, his)

		item := new(endpoints.Story)
		item.ID, _ = strconv.Atoi(id)
		item.Title = messages.EnterCommentSectionToUpdate
		item.Time = time.Now().Unix()
		item.Author = "[]"

		_ = sh.AddItemToFavoritesAndWriteToFile(item)

		println("Item added to favorites")
		println("(enter the comment section from clx to update)")
	},
}
