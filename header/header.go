package header

import (
	"clx/screen"
	"clx/utils/formatter"
	"strings"
)

const (
	leftPadding         = "    "
	symbolHeaderSpacing = "  "
)

func GetHackerNewsHeader(selectedSubHeader int, showFavorites bool, orangeHeader bool) string {
	if showFavorites {
		return header("🆈", "Hacker News   ", []string{"new", "ask", "show", "favorites"},
			selectedSubHeader, orangeHeader)
	}

	return header("🆈", "Hacker News   ", []string{"new", "ask", "show"},
		selectedSubHeader, orangeHeader)
}

func GetCircumflexHeader() string {
	return ""
}

func header(symbol string, title string, subHeaders []string, selectedSubHeader int, orangeHeader bool) string {
	background := getBackground(orangeHeader)
	screenWidth := screen.GetTerminalWidth()

	titleHeader := background + leftPadding + symbol + symbolHeaderSpacing + title
	categoryHeader := getCategoryHeader(subHeaders, selectedSubHeader)
	whitespaceFiller := getWhitespaceFiller(titleHeader+categoryHeader, screenWidth)

	return titleHeader + categoryHeader + whitespaceFiller
}

func getBackground(orangeHeader bool) string {
	if orangeHeader {
		return "[#0c0c0c:#FFA500:bu]"
	}

	return "[::bu]"
}

func getWhitespaceFiller(base string, screenWidth int) string {
	availableScreenSpace := screenWidth - formatter.Len(base)

	if availableScreenSpace < 0 {
		return ""
	}

	return strings.Repeat(" ", availableScreenSpace)
}

func getCategoryHeader(subHeaders []string, selectedSubHeader int) string {
	formattedCategory := ""
	itemsTotal := len(subHeaders)
	selectedOpen := "[::rb]"
	selectedClose := "[::bu]"

	for i, subHeader := range subHeaders {
		isOnLastItem := i == itemsTotal-1
		separator := getSeparator(isOnLastItem)

		if i+1 == selectedSubHeader {
			formattedCategory += selectedOpen + subHeader + selectedClose + separator
		} else {
			formattedCategory += subHeader + separator
		}
	}

	return formattedCategory
}

func getSeparator(isOnLastItem bool) string {
	if isOnLastItem {
		return ""
	}

	return " | "
}
