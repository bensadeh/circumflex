package bheader

import (
	"github.com/muesli/termenv"
)

const (
	gray      = "237"
	lightGray = "238"
	magenta   = "200"
	yellow    = "214"
	blue      = "69"
)

func GetHeader(selectedSubHeader int, showFavorites bool, width int) string {
	categories := getCategories(showFavorites)

	return header(categories, selectedSubHeader)

}

func getCategories(showFavorites bool) []string {
	if showFavorites {
		return []string{"new", "ask", "show", "favorites"}
	}

	return []string{"new", "ask", "show"}
}

func header(subHeaders []string, selectedSubHeader int) string {
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

	//titleStyle := lipgloss.NewStyle().
	//	Background()
	//title := background1() + "  " + magenta() + "c" + yellow() + "l" + blue() + "x  "
	//
	//background := "[#0c0c0c:navy:-]"
	//black := background2() + categoriesColor()
	//screenWidth := screen.GetTerminalWidth()
	//
	//titleOpenTag := "[:navy:]"
	//titleCloseTag := "[:navy:-]"
	//formattedTitle := titleOpenTag + title + titleCloseTag
	//
	//titleHeader := background + formattedTitle + black + "  "
	//categoryHeader := getCategoryHeaderDark(subHeaders, selectedSubHeader)
	//whitespaceFiller := getWhitespaceFillerNew(titleHeader+categoryHeader, screenWidth)

	return title
}
