package style

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	magentaDark = "5"
	yellowDark  = "3"
	blueDark    = "4"
	pinkDark    = "219"

	orange      = "214"
	orangeFaint = "94"

	logoBgDark           = "0"
	headerBgDark         = "8"
	unselectedItemFgDark = "251"
	paginatorBgDark      = logoBgDark
	selectedPageFgDark   = unselectedItemFgDark
	unselectedPageFgDark = "239"

	magentaLight = magentaDark
	yellowLight  = "208"
	blueLight    = blueDark
	pinkLight    = pinkDark

	logoBgLight           = "7"
	headerBgLight         = "15"
	unselectedItemFgLight = "235"
	paginatorBgLight      = logoBgLight
	selectedPageFgLight   = unselectedItemFgLight
	unselectedPageFgLight = "247"
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

func GetUnselectedItemFg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{}
}
