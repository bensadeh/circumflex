package bheader

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"strings"
)

const (
	gray       = "237"
	lightGray  = "238"
	magenta    = "200"
	yellow     = "214"
	blue       = "33"
	red        = "219"
	unselected = "250"

	new       = 1
	ask       = 2
	show      = 3
	favorites = 4
)

func GetHeader(selectedSubHeader int, showFavorites bool, width int) string {
	categories := []string{"new", "ask", "show"}

	if showFavorites {
		categories = append(categories, "favorites")
	}

	return header(categories, selectedSubHeader, width)

}

func header(subHeaders []string, selectedSubHeader int, width int) string {
	p := termenv.ColorProfile()
	c := termenv.String("  c").
		Foreground(p.Color(magenta)).
		Background(p.Color(gray))

	l := termenv.String("l").
		Foreground(p.Color(yellow)).
		Background(p.Color(gray))

	x := termenv.String("x  ").
		Foreground(p.Color(blue)).
		Background(p.Color(gray))

	title := c.String() + l.String() + x.String()
	categories := getCategories(subHeaders, selectedSubHeader)
	filler := getFiller(title, categories, width)

	return title + categories + filler
}

func getFiller(title string, categories string, width int) string {
	p := termenv.ColorProfile()
	availableSpace := width - lipgloss.Width(title+categories)

	if availableSpace < 0 {
		return ""
	}

	filler := strings.Repeat(" ", availableSpace)

	return termenv.String(filler).
		Background(p.Color(lightGray)).
		String()
}

func getCategories(subHeaders []string, selectedSubHeader int) string {
	p := termenv.ColorProfile()
	categories := termenv.String("  ").
		Background(p.Color(lightGray)).
		String()

	separator := termenv.String(" • ").
		Foreground(p.Color(unselected)).
		Background(p.Color(lightGray)).
		String()

	for i, subHeader := range subHeaders {
		isOnLastItem := i == len(subHeaders)-1
		selectedCatColor := getColor(i, selectedSubHeader)

		categories += termenv.String(subHeader).
			Foreground(p.Color(selectedCatColor)).
			Background(p.Color(lightGray)).
			String()

		if !isOnLastItem {
			categories += separator
		}

	}

	return categories
}

func getColor(i int, selectedSubHeader int) string {
	if i+1 == selectedSubHeader {
		return getSelectedCategoryColor(i)
	}

	return unselected
}

func getSelectedCategoryColor(selectedSubHeader int) string {
	switch selectedSubHeader {
	case new:
		return magenta
	case ask:
		return yellow
	case show:
		return blue
	case favorites:
		return red
	default:
		return unselected
	}
}
