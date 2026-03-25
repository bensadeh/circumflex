package header

import (
	"clx/categories"
	"clx/style"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

func Header(allCategories []categories.Category, selectedSubHeader int, width int) string {
	c := lipgloss.NewStyle().Foreground(style.HeaderC())
	l := lipgloss.NewStyle().Foreground(style.HeaderL())
	x := lipgloss.NewStyle().Foreground(style.HeaderX())

	title := c.Render("  c") + l.Render("l") + x.Render("x  ")
	cats := getCategories(allCategories, selectedSubHeader)
	filler := getFiller(title, cats, width)

	return title + cats + filler + "\n" + strings.Repeat("‾", width)
}

func getFiller(title string, categories string, width int) string {
	availableSpace := width - lipgloss.Width(title+categories)

	if availableSpace < 0 {
		return ""
	}

	return strings.Repeat(" ", availableSpace)
}

func getCategories(allCategories []categories.Category, selectedSubHeader int) string {
	cats := allCategories[1:]

	var out strings.Builder

	separator := lipgloss.NewStyle().
		Faint(true).
		Render(" • ")

	for i, cat := range cats {
		name := categories.Name(cat)
		selectedCatColor, isSelected := getColor(i+1, selectedSubHeader, cat)

		out.WriteString(lipgloss.NewStyle().
			Foreground(selectedCatColor).
			Faint(!isSelected).
			Render(name))

		if i < len(cats)-1 {
			out.WriteString(separator)
		}
	}

	return out.String()
}

func getColor(index int, selectedSubHeader int, cat categories.Category) (clr color.Color, isSelected bool) {
	if index == selectedSubHeader {
		return getSelectedCategoryColor(selectedSubHeader, cat)
	}

	return lipgloss.NoColor{}, false
}

func getSelectedCategoryColor(selectedSubHeader int, cat categories.Category) (clr color.Color, isSelected bool) {
	if cat == categories.Favorites {
		return style.HeaderFavorites(), true
	}

	primary := style.HeaderPrimary()
	secondary := style.HeaderSecondary()
	tertiary := style.HeaderTertiary()

	switch selectedSubHeader % 3 {
	case 0:
		return tertiary, true
	case 1:
		return primary, true
	case 2:
		return secondary, true
	}

	return tertiary, true
}
