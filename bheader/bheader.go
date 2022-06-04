package bheader

import (
	"clx/constants/category"
	"clx/constants/style"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

func GetHeader(selectedSubHeader int, width int) string {
	bg := style.GetLogoBg()

	c := lipgloss.NewStyle().
		Foreground(style.GetMagenta()).
		Background(bg)

	l := lipgloss.NewStyle().
		Foreground(style.GetYellow()).
		Background(bg)

	x := lipgloss.NewStyle().
		Foreground(style.GetBlue()).
		Background(bg)

	title := c.Render("  c") + l.Render("l") + x.Render("x  ")

	categories := getCategories(selectedSubHeader)
	filler := getFiller(title, categories, width)

	return title + categories + filler
}

func getFiller(title string, categories string, width int) string {
	availableSpace := width - lipgloss.Width(title+categories)

	if availableSpace < 0 {
		return ""
	}

	filler := strings.Repeat(" ", availableSpace)

	return lipgloss.NewStyle().
		Background(style.GetHeaderBg()).
		Render(filler)
}

func getCategories(selectedSubHeader int) string {
	subHeaders := []string{"new", "ask", "show"}
	fg := style.GetUnselectedItemFg()
	bg := style.GetHeaderBg()

	categories := lipgloss.NewStyle().
		Background(bg).
		Render("   ")

	separator := lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Render(" • ")

	for i, subHeader := range subHeaders {
		isOnLastItem := i == len(subHeaders)-1
		selectedCatColor, isSelected := getColor(i, selectedSubHeader)

		categories += lipgloss.NewStyle().
			Foreground(selectedCatColor).
			Background(bg).
			Bold(isSelected).
			Render(subHeader)

		if !isOnLastItem {
			categories += separator
		}
	}

	return categories
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
