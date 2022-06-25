package style

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const (
	magentaDark = "200"
	yellowDark  = "214"
	blueDark    = "33"
	pinkDark    = "219"

	black       = "16"
	orange      = "214"
	orangeFaint = "94"

	logoBgDark              = "#0f1429"
	labelFgDark             = yellowDark
	labelMarkAsReadFgDark   = yellowDark
	labelBgDark             = logoBgDark
	labelMarkAsReadBgDark   = headerBgDark
	headerBgDark            = "#2d3454"
	unselectedItemFgDark    = "247"
	paginatorBgDark         = logoBgDark
	selectedPageFgDark      = unselectedItemFgDark
	unselectedPageFgDark    = "239"
	ycLogoFgDark            = orange
	ycLogoMarkAsReadFgDark  = orangeFaint
	ycLabelBgDark           = logoBgDark
	ycLabelMarkAsReadBgDark = orangeFaint
	ycTextFgDark            = unselectedItemFgDark
	ycTextMarkAsReadFgDark  = ycTextFgDark

	magentaLight = magentaDark
	yellowLight  = "208"
	blueLight    = blueDark
	pinkLight    = pinkDark

	logoBgLight              = "252"
	labelFgLight             = yellowLight
	labelMarkAsReadFgLight   = yellowLight
	labelBgLight             = headerBgLight
	labelMarkAsReadBgLight   = headerBgLight
	headerBgLight            = "254"
	unselectedItemFgLight    = "235"
	paginatorBgLight         = logoBgLight
	selectedPageFgLight      = unselectedItemFgLight
	unselectedPageFgLight    = "247"
	ycLogoFgLight            = black
	ycLogoMarkAsReadFgLight  = "245"
	ycLabelBgLight           = orange
	ycLabelMarkAsReadBgLight = "253"
	ycTextFgLight            = unselectedItemFgLight
	ycTextMarkAsReadFgLight  = "245"
)

func GetMagenta() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: magentaLight, Dark: magentaDark}
}

func GetYellow() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: yellowLight, Dark: yellowDark}
}

func GetBlue() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: blueLight, Dark: blueDark}
}

func GetPink() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: pinkLight, Dark: pinkDark}
}

func GetOrange() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: orange, Dark: orange}
}

func GetOrangeFaint() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: orangeFaint, Dark: orangeFaint}
}

func GetLogoBg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: logoBgLight, Dark: logoBgDark}
}

func GetHeaderBg() lipgloss.TerminalColor {
	profile := termenv.ColorProfile()

	if profile != termenv.TrueColor {
		return lipgloss.AdaptiveColor{Light: headerBgLight, Dark: "237"}
	}

	return lipgloss.AdaptiveColor{Light: headerBgLight, Dark: headerBgDark}
}

func GetStatusBarBg() lipgloss.TerminalColor {
	return GetHeaderBg()
}

func GetPaginatorBg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: paginatorBgLight, Dark: paginatorBgDark}
}

func GetUnselectedItemFg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: unselectedItemFgLight, Dark: unselectedItemFgDark}
}

func GetSelectedPageFg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: selectedPageFgLight, Dark: selectedPageFgDark}
}

func GetUnselectedPageFg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: unselectedPageFgLight, Dark: unselectedPageFgDark}
}
