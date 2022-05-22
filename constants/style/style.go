package style

import "github.com/charmbracelet/lipgloss"

const (
	magentaDark = "200"
	yellowDark  = "214"
	blueDark    = "33"
	pinkDark    = "219"

	logoBackgroundDark   = "234"
	headerBackgroundDark = "235"
	unselectedItemDark   = "250"
	selectedPageDark     = unselectedItemDark
	unselectedPageDark   = "239"

	magentaLight = magentaDark
	yellowLight  = "208"
	blueLight    = blueDark
	pinkLight    = pinkDark

	logoBackgroundLight   = "254"
	headerBackgroundLight = "247"
	unselectedItemLight   = "235"
	selectedPageLight     = unselectedItemLight
	unselectedPageLight   = "247"
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

func GetLogoBackground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: logoBackgroundLight, Dark: logoBackgroundDark}
}

func GetHeaderBackground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: headerBackgroundLight, Dark: headerBackgroundDark}
}

func GetUnselectedItemForeground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: unselectedItemLight, Dark: unselectedItemDark}
}

func GetSelectedPageForeground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: selectedPageLight, Dark: selectedPageDark}
}

func GetUnselectedPageForeground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: unselectedPageLight, Dark: unselectedPageDark}
}
