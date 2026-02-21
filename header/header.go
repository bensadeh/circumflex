package header

import (
	"clx/categories"
	"fmt"
	"strings"

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

	var cats strings.Builder
	cats.WriteString(lipgloss.NewStyle().
		Underline(true).
		Render(""))

	separator := lipgloss.NewStyle().
		Faint(true).
		Render(" • ")

	for i, subHeader := range subHeaders {
		isOnLastItem := i == len(subHeaders)-1
		selectedCatColor, isSelected := getColor(i, selectedSubHeader, len(subHeaders), hasFavorites)

		cats.WriteString(lipgloss.NewStyle().
			Foreground(selectedCatColor).
			Faint(!isSelected).
			Render(subHeader))

		if !isOnLastItem {
			cats.WriteString(separator)
		}
	}

	return cats.String()
}

var ReverseCategoryMapping = map[int]categories.Category{
	0: categories.Top,
	1: categories.Newest,
	2: categories.Ask,
	3: categories.Show,
	4: categories.Best,
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
		return []string{}
	}

	return list[1:]
}

func getColor(i int, selectedSubHeader int, numCategories int, hasFavorites bool) (color lipgloss.TerminalColor, isSelected bool) {
	if i+1 == selectedSubHeader {
		return getSelectedCategoryColor(i+1, numCategories, hasFavorites)
	}

	return lipgloss.NoColor{}, false
}

func getSelectedCategoryColor(selectedSubHeader int, numCategories int, hasFavorites bool) (color lipgloss.TerminalColor,
	isSelected bool) {
	magenta := lipgloss.Color("5")
	yellow := lipgloss.Color("3")
	blue := lipgloss.Color("4")
	pink := lipgloss.Color("219")

	if hasFavorites && selectedSubHeader == numCategories {
		return pink, true
	}

	switch selectedSubHeader % 3 {
	case 0:
		return blue, true
	case 1:
		return magenta, true
	case 2:
		return yellow, true
	default:
		panic("unreachable code")
	}
}
