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

func GetHackerNewsHeader(selectedSubHeader int, showFavorites bool, headerType int) string {
	categories := getCategories(showFavorites)

	switch headerType {
	case 1:
		return header(getSymbol(false), "Hacker News   ", categories, selectedSubHeader, false)

	case 2:
		return header(getSymbol(true), "Hacker News   ", categories, selectedSubHeader, true)

	default:
		return headerNew(categories, selectedSubHeader)
	}
}

func getSymbol(orangeHeader bool) string {
	if orangeHeader {
		return "🅈"
	}

	return "🆈"
}

func getCategories(showFavorites bool) []string {
	if showFavorites {
		return []string{"new", "ask", "show", "favorites"}
	}

	return []string{"new", "ask", "show"}
}

func header(symbol string, title string, subHeaders []string, selectedSubHeader int, orangeHeader bool) string {
	background := getBackground(orangeHeader)
	screenWidth := screen.GetTerminalWidth()

	symbolOpenTag := getSymbolOpenTag(orangeHeader)
	symbolCloseTag := background
	formattedSymbol := symbolOpenTag + symbol + symbolCloseTag

	titleOpenTag := getTitleOpenTag(orangeHeader)
	titleCloseTag := getTitleCloseTag(orangeHeader)
	formattedTitle := titleOpenTag + title + titleCloseTag

	titleHeader := background + leftPadding + formattedSymbol + symbolHeaderSpacing + formattedTitle + background
	categoryHeader := getCategoryHeader(subHeaders, selectedSubHeader, orangeHeader)
	whitespaceFiller := getWhitespaceFiller(titleHeader+categoryHeader, screenWidth)

	return titleHeader + categoryHeader + whitespaceFiller
}

func getBackground(orangeHeader bool) string {
	if orangeHeader {
		return "[#0c0c0c:#ff6600:-]"
	}

	return "[::bu]"
}

func getSymbolOpenTag(orangeHeader bool) string {
	if orangeHeader {
		return "[#FFFFFF:#ff6600:b]"
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

func getCategoryHeader(subHeaders []string, selectedSubHeader int, orangeHeader bool) string {
	formattedCategory := ""
	itemsTotal := len(subHeaders)
	selectedOpen := getSelectedCategoryOpenTag(orangeHeader)
	selectedClose := getSelectedCategoryCloseTag(orangeHeader)

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

func getTitleOpenTag(orangeHeader bool) string {
	if orangeHeader {
		return "[::b]"
	}

	return "[::bu]"
}

func getTitleCloseTag(orangeHeader bool) string {
	if orangeHeader {
		return "[::-]"
	}

	return "[::bu]"
}

func getSelectedCategoryOpenTag(orangeHeader bool) string {
	if orangeHeader {
		return "[::r]"
	}

	return "[::rb]"
}

func getSelectedCategoryCloseTag(orangeHeader bool) string {
	if orangeHeader {
		return "[::-]"
	}

	return "[::bu]"
}

func getSeparator(isOnLastItem bool) string {
	if isOnLastItem {
		return ""
	}

	return " | "
}
