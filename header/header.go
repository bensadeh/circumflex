package header

import (
	"clx/screen"
	"clx/utils/format"
	"strings"
)

const (
	leftPadding         = "    "
	symbolHeaderSpacing = "  "
)

func GetHackerNewsHeader(selectedSubHeader int, showFavorites bool) string {
	if showFavorites {
		return header("🆈", "Hacker News   ", []string{"new", "ask", "show", "favorites"}, selectedSubHeader)
	}

	return header("🆈", "Hacker News   ", []string{"new", "ask", "show"}, selectedSubHeader)
}

func GetCircumflexHeader(selectedSubHeader int) string {
	return header("🅲", "circumflex    ", []string{"keymaps", "settings"}, selectedSubHeader)
}

func header(symbol string, title string, subHeaders []string, selectedSubHeader int) string {
	background := "[-::bu]"
	screenWidth := screen.GetTerminalWidth()

	titleHeader := background + leftPadding + symbol + symbolHeaderSpacing + title
	categoryHeader := getCategoryHeader(subHeaders, selectedSubHeader)
	whitespaceFiller := getWhitespaceFiller(titleHeader+categoryHeader, screenWidth)

	return titleHeader + categoryHeader + whitespaceFiller
}

func getWhitespaceFiller(base string, screenWidth int) string {
	availableScreenSpace := screenWidth - format.Len(base)

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
