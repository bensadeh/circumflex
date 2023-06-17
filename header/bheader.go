package header

import (
	"clx/categories"
	"fmt"
	"strings"

	"clx/constants/category"
	"github.com/charmbracelet/lipgloss"
)

func GetHeader(allCategories []int, hasFavorites bool, selectedSubHeader int, width int) string {
	c := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	l := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	x := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))

	title := c.Render("  c") + l.Render("l") + x.Render("x  ")
	cats := getCategories(allCategories, hasFavorites, selectedSubHeader)
	filler := getFiller(title, cats, width)

	return title + cats + filler + "\n" + strings.Repeat("‾", width)
}

func getFiller(title string, categories string, width int) string {
	availableSpace := width - lipgloss.Width(title+categories)

	if availableSpace < 0 {
		return ""
	}

	filler := strings.Repeat(" ", availableSpace)

	return lipgloss.NewStyle().Render(filler)
}

func getCategories(allCategories []int, hasFavorites bool, selectedSubHeader int) string {
	subHeaders := getSubHeaders(allCategories, hasFavorites)
	subHeaders = removeFirstElement(subHeaders)

	cats := lipgloss.NewStyle().
		Underline(true).
		Render("")

	separator := lipgloss.NewStyle().
		Faint(true).
		Render(" • ")

	for i, subHeader := range subHeaders {
		isOnLastItem := i == len(subHeaders)-1
		selectedCatColor, isSelected := getColor(i, selectedSubHeader)

		cats += lipgloss.NewStyle().
			Foreground(selectedCatColor).
			Faint(!isSelected).
			Render(subHeader)

		if !isOnLastItem {
			cats += separator
		}
	}

	return cats
}

var ReverseCategoryMapping = map[int]categories.Category{
	0: categories.FrontPage,
	1: categories.Newest,
	2: categories.Ask,
	3: categories.Show,
	4: categories.Favorites,
}

func getSubHeaders(allCategories []int, hasFavorites bool) []string {
	var cats []string
	for _, cat := range allCategories {
		if categoryName, exists := ReverseCategoryMapping[cat]; exists {
			cats = append(cats, string(categoryName))
		} else {
			panic(fmt.Sprintf("Invalid category ID: %d", cat))
		}
	}

	if hasFavorites {
		cats = append(cats, "favorites")
	}

	return cats
}

func removeFirstElement(list []string) []string {
	if len(list) == 0 {
		return []string{} // return an empty slice when the input list is empty
	}

	// return a slice starting from the second element (index 1) to the end of the list
	// this will not mutate the original list, it will create a new one
	return list[1:]
}

//func getSubHeaders(allCategories []int) []string {
//	if favoritesHasItems {
//		return []string{"new", "ask", "show", "favorites"}
//	}
//
//	return []string{"new", "ask", "show"}
//}

func getColor(i int, selectedSubHeader int) (color lipgloss.TerminalColor, isSelected bool) {
	if i+1 == selectedSubHeader {
		return getSelectedCategoryColor(i + 1)
	}

	return lipgloss.NoColor{}, false
}

func getSelectedCategoryColor(selectedSubHeader int) (color lipgloss.TerminalColor, isSelected bool) {
	magenta := lipgloss.Color("5")
	yellow := lipgloss.Color("3")
	blue := lipgloss.Color("4")
	pink := lipgloss.Color("219")

	switch selectedSubHeader {
	case category.New:
		return magenta, true
	case category.Ask:
		return yellow, true
	case category.Show:
		return blue, true
	case category.Favorites:
		return pink, true
	default:
		panic("unsupported header category")
	}
}
