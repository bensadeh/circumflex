package header

import (
	"clx/screen"
	"clx/utils/format"
)

const (
	leftPadding         = "    "
	symbolHeaderSpacing = "  "
)

func GetHackerNewsHeader(selectedSubHeader int) string {
	return header("🆈", "Hacker News   ", []string{"new", "ask", "show"}, selectedSubHeader, "orange")
}

func GetCircumflexHeader(selectedSubHeader int) string {
	return header("🅲", "circumflex    ", []string{"keymaps", "settings"}, selectedSubHeader, "#82aaff")
}

func header(symbol string, title string, subHeaders []string, selectedSubHeader int, bgColor string) string {
	background := getBackground(bgColor)
	screenWidth := screen.GetTerminalWidth()

	mainHeader := background + blackInBold() + leftPadding + symbol + symbolHeaderSpacing + title + format.ResetStyle()
	subHeader := getSubHeader(subHeaders, selectedSubHeader)
	whitespaceFiller := getWhitespaceFiller(mainHeader+subHeader, screenWidth)

	return mainHeader + subHeader + whitespaceFiller
}

func getWhitespaceFiller(base string, screenWidth int) string {
	availableScreenSpace := screenWidth - format.Len(base)
	whitespace := ""

	for i := 0; i < availableScreenSpace; i++ {
		whitespace += " "
	}

	return whitespace
}

func getSubHeader(subHeaders []string, selectedSubHeader int) string {
	formattedSubHeader := ""
	itemsTotal := len(subHeaders)

	for i, subHeader := range subHeaders {
		isOnLastItem := i == itemsTotal-1
		separator := getSeparator(isOnLastItem)

		if i+1 == selectedSubHeader {
			formattedSubHeader += format.White(subHeader) + format.BlackNoReset("") + separator
		} else {
			formattedSubHeader += subHeader + separator
		}
	}

	return formattedSubHeader
}

func getSeparator(isOnLastItem bool) string {
	if isOnLastItem {
		return ""
	}

	return " | "
}

func blackInBold() string {
	return "[#0c0c0c::b]"
}

func getBackground(color string) string {
	return "[:" + color + ":]"
}
