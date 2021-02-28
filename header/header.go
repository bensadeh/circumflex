package header

import (
	"clx/screen"
	"clx/utils/format"
	"strings"
)

const (
	leftPadding         = "    "
	symbolHeaderSpacing = "  "
	palenightBlue       = "#82aaff"
	black               = "#0c0c0c"
)

func GetHackerNewsHeader(selectedSubHeader int) string {
	return header("🆈", "Hacker News   ", []string{"new", "ask", "show"},
		selectedSubHeader, "", "-", true)
}

func GetCircumflexHeader(selectedSubHeader int) string {
	return header("🅲", "circumflex    ", []string{"keymaps", "settings"},
		selectedSubHeader, palenightBlue, black, false)
}

func header(symbol string, title string, subHeaders []string, selectedSubHeader int,
	bgColor, fgColor string, isTransparentHeader bool) string {
	background := getBackground(bgColor, isTransparentHeader)
	screenWidth := screen.GetTerminalWidth()

	titleHeader := background + fgColorAndBold(fgColor, isTransparentHeader) + leftPadding + symbol + symbolHeaderSpacing +
		title + getTitleHeaderSeparator(isTransparentHeader)
	categoryHeader := getCategoryHeader(subHeaders, selectedSubHeader, fgColor, "white",
		isTransparentHeader)
	whitespaceFiller := getWhitespaceFiller(titleHeader+categoryHeader, screenWidth)

	return titleHeader + categoryHeader + whitespaceFiller
}

func getTitleHeaderSeparator(isTransparentHeader bool) string {
	if isTransparentHeader {
		return ""
	}

	return format.ResetStyle()
}

func getWhitespaceFiller(base string, screenWidth int) string {
	availableScreenSpace := screenWidth - format.Len(base)

	return strings.Repeat(" ", availableScreenSpace)
}

func getCategoryHeader(subHeaders []string, selectedSubHeader int, fgColor, selectedColor string,
	isTransparentTopBar bool) string {
	formattedCategory := ""
	itemsTotal := len(subHeaders)
	selectedOpen := getSelectedMarkerOpen(selectedColor, isTransparentTopBar)
	selectedClose := getSelectedMarkerClosed(fgColor, isTransparentTopBar)

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

func getSelectedMarkerClosed(color string, isTransparentTopBar bool) string {
	if isTransparentTopBar {
		return "[" + color + "::bu]"
	}

	return "[" + color + "::-]"
}

func getSelectedMarkerOpen(color string, isTransparentTopBar bool) string {
	if isTransparentTopBar {
		return "[::rbu]"
	}

	return "[" + color + "::]"
}

func getSeparator(isOnLastItem bool) string {
	if isOnLastItem {
		return ""
	}

	return " | "
}

func fgColorAndBold(fgColor string, isTransparentHeader bool) string {
	if isTransparentHeader {
		return "[" + fgColor + "::bu]"
	}

	return "[" + fgColor + "::b]"
}

func getBackground(color string, isTransparentHeader bool) string {
	if isTransparentHeader {
		return "[::u]"
	}

	return "[:" + color + ":]"
}
