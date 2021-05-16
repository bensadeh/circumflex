package info

import (
	"clx/constants/categories"
	"clx/constants/messages"
	"clx/keymaps"

	"github.com/gdamore/tcell/v2"
)

const (
	numberOfCategories = 3
)

func GetStatusBarText(category int) string {
	if category == categories.Definition {
		return messages.GetCircumflexStatusMessage()
	}

	return ""
}

func GetNewCategory(event *tcell.EventKey, currentCategory int) int {
	if event.Key() == tcell.KeyBacktab {
		return getPreviousCategory(currentCategory)
	}

	return getNextCategory(currentCategory)
}

func getNextCategory(currentCategory int) int {
	isOnLastCategory := currentCategory == (numberOfCategories - 1)

	if isOnLastCategory {
		return 0
	}

	return currentCategory + 1
}

func getPreviousCategory(currentCategory int) int {
	isOnFirstCategory := currentCategory == 0

	if isOnFirstCategory {
		return numberOfCategories - 1
	}

	return currentCategory - 1
}

func GetText(category int, screenWidth int) string {
	switch category {
	case categories.Definition:
		return getKeymaps(screenWidth)

	case categories.Keymaps:
		return getKeymaps(screenWidth)

	case categories.Settings:
		return getKeymaps(screenWidth)

	default:
		return ""
	}
}

func getKeymaps(screenWidth int) string {
	keys := new(keymaps.List)
	keys.Init()

	keys.AddSeparator()
	keys.AddHeader("Main View")
	keys.AddSeparator()
	keys.AddKeymap("Read comments", "Enter")
	keys.AddKeymap("Read article in Reader Mode", "Space")
	keys.AddKeymap("Change category", "Tab")

	keys.AddKeymap("Open story link in browser", "o")
	keys.AddKeymap("Open comments in browser", "c")
	keys.AddKeymap("Refresh", "r")
	keys.AddSeparator()
	keys.AddKeymap("Add to favorites", "f")
	keys.AddKeymap("Add to favorites by ID", "F")
	keys.AddKeymap("Delete from favorites", "x")
	keys.AddSeparator()
	keys.AddKeymap("Bring up this screen", "i, ?")
	keys.AddKeymap("Quit to prompt", "q")
	keys.AddSeparator()
	keys.AddHeader("Comment Section")
	keys.AddSeparator()
	keys.AddKeymap("Down one half-window", "d")
	keys.AddKeymap("Up one half-window", "u")
	keys.AddSeparator()
	keys.AddKeymap("Jump to next top-level comment", "/ + '::'")
	keys.AddKeymap("Repeat last search", "n")
	keys.AddKeymap("Repeat last search in reverse direction", "N")
	keys.AddSeparator()
	keys.AddKeymap("Help screen", "h")
	keys.AddKeymap("Quit to Main Screen", "q")

	return keys.Print(screenWidth)
}
