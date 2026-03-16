package header

import (
	"clx/categories"
	"clx/style"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

func GetHeader(allCategories []int, selectedSubHeader int, width int) string {
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

	filler := strings.Repeat(" ", availableSpace)

	return lipgloss.NewStyle().Render(filler)
}

func getCategories(allCategories []int, selectedSubHeader int) string {
	subHeaders := getSubHeaders(allCategories)
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
		selectedCatColor, isSelected := getColor(i, selectedSubHeader, allCategories)

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

func getSubHeaders(allCategories []int) []string {
	var cats []string
	for _, cat := range allCategories {
		cats = append(cats, categories.Name(cat))
	}

	return cats
}

func removeFirstElement(list []string) []string {
	if len(list) == 0 {
		return []string{}
	}

	return list[1:]
}

func getColor(i int, selectedSubHeader int, allCategories []int) (clr color.Color, isSelected bool) {
	if i+1 == selectedSubHeader {
		return getSelectedCategoryColor(selectedSubHeader, allCategories[i+1])
	}

	return lipgloss.NoColor{}, false
}

func getSelectedCategoryColor(selectedSubHeader int, cat int) (clr color.Color, isSelected bool) {
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
	default:
		return tertiary, true
	}
}
