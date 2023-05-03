package header

import (
	"strings"

	"clx/constants/category"
	"clx/constants/style"
	"github.com/charmbracelet/lipgloss"
)

func GetHeader(selectedSubHeader int, favoritesHasItems bool, width int) string {
	c := lipgloss.NewStyle().Foreground(style.GetMagenta())
	l := lipgloss.NewStyle().Foreground(style.GetYellow())
	x := lipgloss.NewStyle().Foreground(style.GetBlue())

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

	return lipgloss.NewStyle().
		//Background(style.GetHeaderBg()).
		Render(filler)
}

func getCategories(selectedSubHeader int, favoritesHasItems bool) string {
	subHeaders := getSubHeaders(favoritesHasItems)
	fg := style.GetUnselectedItemFg()
	//bg := style.GetHeaderBg()

	categories := lipgloss.NewStyle().
		//Background(bg).
		Underline(true).
		Render("")

	separator := lipgloss.NewStyle().
		Foreground(fg).
		Faint(true).
		//Background(bg).
		Render(" • ")

	for i, subHeader := range subHeaders {
		isOnLastItem := i == len(subHeaders)-1
		selectedCatColor, isSelected := getColor(i, selectedSubHeader)

		categories += lipgloss.NewStyle().
			Foreground(selectedCatColor).
			Faint(!isSelected).
			//Background(bg).
			//Bold(isSelected).
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

func getColor(i int, selectedSubHeader int) (lipgloss.TerminalColor, bool) {
	if i+1 == selectedSubHeader {
		return getSelectedCategoryColor(i + 1)
	}

	return style.GetUnselectedItemFg(), false
}

func getSelectedCategoryColor(selectedSubHeader int) (lipgloss.TerminalColor, bool) {
	switch selectedSubHeader {
	case category.New:
		return style.GetMagenta(), true
	case category.Ask:
		return style.GetYellow(), true
	case category.Show:
		return style.GetBlue(), true
	case category.Favorites:
		return style.GetPink(), true
	default:
		return style.GetUnselectedItemFg(), false
	}
}
