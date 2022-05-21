package bheader

import (
	"clx/constants/category"
	"clx/constants/style"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

func GetHeader(selectedSubHeader int, width int) string {
	bg := lipgloss.AdaptiveColor{Light: style.LogoBackgroundLight, Dark: style.LogoBackgroundDark}

	c := lipgloss.NewStyle().
		Foreground(lipgloss.Color(style.Magenta)).
		Background(bg)

	l := lipgloss.NewStyle().
		Foreground(lipgloss.Color(style.Yellow)).
		Background(bg)

	x := lipgloss.NewStyle().
		Foreground(lipgloss.Color(style.Blue)).
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
		Background(lipgloss.AdaptiveColor{Light: style.HeaderBackgroundLight, Dark: style.HeaderBackgroundDark}).
		Render(filler)
}

func getCategories(selectedSubHeader int) string {
	subHeaders := []string{"new", "ask", "show"}
	fg := lipgloss.AdaptiveColor{Light: style.UnselectedItemLight, Dark: style.UnselectedItemDark}
	bg := lipgloss.AdaptiveColor{Light: style.HeaderBackgroundLight, Dark: style.HeaderBackgroundDark}

	categories := lipgloss.NewStyle().
		Background(bg).
		Render("   ")

	separator := lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Render(" • ")

	for i, subHeader := range subHeaders {
		isOnLastItem := i == len(subHeaders)-1
		selectedCatColor := getColor(i, selectedSubHeader)

		categories += lipgloss.NewStyle().
			Foreground(selectedCatColor).
			Background(bg).
			Render(subHeader)

		if !isOnLastItem {
			categories += separator
		}

	}

	return categories
}

func getColor(i int, selectedSubHeader int) lipgloss.TerminalColor {
	if i+1 == selectedSubHeader {
		return getSelectedCategoryColor(i + 1)
	}

	return lipgloss.AdaptiveColor{Light: style.UnselectedItemLight, Dark: style.UnselectedItemDark}
}

func getSelectedCategoryColor(selectedSubHeader int) lipgloss.TerminalColor {
	switch selectedSubHeader {
	case category.New:
		return lipgloss.Color(style.Magenta)
	case category.Ask:
		return lipgloss.Color(style.Yellow)
	case category.Show:
		return lipgloss.Color(style.Blue)
	case category.Favorites:
		return lipgloss.Color(style.Red)
	default:
		return lipgloss.AdaptiveColor{Light: style.UnselectedItemLight, Dark: style.UnselectedItemDark}
	}
}
