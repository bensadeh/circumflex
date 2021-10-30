package header

import (
	"clx/screen"
	"clx/utils/formatter"
	"strings"

	"code.rocketnine.space/tslocum/cview"
)

func magenta() string {
	return cview.TranslateANSI("\u001B[38;5;200m")
}

func yellow() string {
	return cview.TranslateANSI("\u001B[38;5;214m")
}

func blue() string {
	return cview.TranslateANSI("\u001B[38;5;69m")
}

func red() string {
	return cview.TranslateANSI("\u001B[38;5;219m")
}

func categoriesColor() string {
	return cview.TranslateANSI("\u001B[38;5;250m")
}

func background1() string {
	return cview.TranslateANSI("\u001b[48;5;237m")
}

func background2() string {
	return cview.TranslateANSI("\u001b[48;5;238m")
}

func headerNew(subHeaders []string, selectedSubHeader int) string {
	title := background1() + "  " + magenta() + "c" + yellow() + "l" + blue() + "x  "

	background := "[#0c0c0c:navy:-]"
	black := background2() + categoriesColor()
	screenWidth := screen.GetTerminalWidth()

	titleOpenTag := "[:navy:]"
	titleCloseTag := "[:navy:-]"
	formattedTitle := titleOpenTag + title + titleCloseTag

	titleHeader := background + formattedTitle + black + "  "
	categoryHeader := getCategoryHeaderNew(subHeaders, selectedSubHeader)
	whitespaceFiller := getWhitespaceFillerNew(titleHeader+categoryHeader, screenWidth)

	return titleHeader + categoryHeader + whitespaceFiller
}

func getWhitespaceFillerNew(base string, screenWidth int) string {
	availableScreenSpace := screenWidth - formatter.Len(base)

	if availableScreenSpace < 0 {
		return ""
	}

	return strings.Repeat(" ", availableScreenSpace)
}

func getCategoryHeaderNew(subHeaders []string, selectedSubHeader int) string {
	formattedCategory := ""
	itemsTotal := len(subHeaders)
	selectedOpen := getSelectedCategoryOpenTagNew(selectedSubHeader)
	selectedClose := categoriesColor()

	for i, subHeader := range subHeaders {
		isOnLastItem := i == itemsTotal-1
		separator := getSeparatorNew(isOnLastItem)

		if i+1 == selectedSubHeader {
			formattedCategory += selectedOpen + subHeader + selectedClose + separator
		} else {
			formattedCategory += subHeader + separator
		}
	}

	return formattedCategory
}

func getSelectedCategoryOpenTagNew(selectedSubHeader int) string {
	switch selectedSubHeader {
	case 1:
		return magenta()
	case 2:
		return yellow()
	case 3:
		return blue()
	case 4:
		return red()
	default:
		return categoriesColor()
	}
}

func getSeparatorNew(isOnLastItem bool) string {
	if isOnLastItem {
		return ""
	}

	return " • "
}
