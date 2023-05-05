package header

import (
	"strings"

	"clx/constants/category"
	"clx/constants/style"
	"github.com/charmbracelet/lipgloss"
)

func GetHeader(selectedSubHeader int, favoritesHasItems bool, width int) string {
	c := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	l := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	x := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))

	title := c.Render("  c") + l.Render("l") + x.Render("x  ")
	categories := getCategories(selectedSubHeader, favoritesHasItems)
	filler := getFiller(title, categories, width)

	return title + categories + filler + "\n" + strings.Repeat("‾", width)
}

func getFiller(title string, categories string, width int) string {
	availableSpace := width - lipgloss.Width(title+categories)

	if availableSpace < 0 {
		return ""
	}

	filler := strings.Repeat(" ", availableSpace)

	return lipgloss.NewStyle().Render(filler)
}

func getCategories(selectedSubHeader int, favoritesHasItems bool) string {
	subHeaders := getSubHeaders(favoritesHasItems)
	fg := style.GetUnselectedItemFg()

	categories := lipgloss.NewStyle().
		Underline(true).
		Render("")

	separator := lipgloss.NewStyle().
		Foreground(fg).
		Faint(true).
		Render(" • ")

	for i, subHeader := range subHeaders {
		isOnLastItem := i == len(subHeaders)-1
		selectedCatColor, isSelected := getColor(i, selectedSubHeader)

		categories += lipgloss.NewStyle().
			Foreground(selectedCatColor).
			Faint(!isSelected).
			Render(subHeader)

		if !isOnLastItem {
			categories += separator
		}
	}

	return categories
}

func getSubHeaders(favoritesHasItems bool) []string {
	if favoritesHasItems {
		return []string{"new", "ask", "show", "favorites"}
	}

	return []string{"new", "ask", "show"}
}

func getColor(i int, selectedSubHeader int) (color lipgloss.TerminalColor, isSelected bool) {
	if i+1 == selectedSubHeader {
		return getSelectedCategoryColor(i + 1)
	}

	return style.GetUnselectedItemFg(), false
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
