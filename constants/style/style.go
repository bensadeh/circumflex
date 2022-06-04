package style

import "github.com/charmbracelet/lipgloss"

const (
	magentaDark = "200"
	yellowDark  = "214"
	blueDark    = "33"
	pinkDark    = "219"

	logoBgDark           = "#292b33"
	headerBgDark         = "#222638"
	unselectedItemFgDark = "247"
	statusBarBgDark      = headerBgDark
	paginatorBgDark      = logoBgDark
	selectedPageFgDark   = unselectedItemFgDark
	unselectedPageFgDark = "239"

	magentaLight = magentaDark
	yellowLight  = "208"
	blueLight    = blueDark
	pinkLight    = pinkDark

	logoBgLight           = "252"
	headerBgLight         = "254"
	unselectedItemFgLight = "235"
	statusBarBgLight      = headerBgLight
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

func GetLogoBackground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: logoBgLight, Dark: logoBgDark}
}

func GetHeaderBackground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: headerBgLight, Dark: headerBgDark}
}

func GetStatusBarBackground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: statusBarBgLight, Dark: statusBarBgDark}
}

func GetPaginatorBackground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: paginatorBgLight, Dark: paginatorBgDark}
}

func GetUnselectedItemForeground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: unselectedItemFgLight, Dark: unselectedItemFgDark}
}

func GetSelectedPageForeground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: selectedPageFgLight, Dark: selectedPageFgDark}
}

func GetUnselectedPageForeground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: unselectedPageFgLight, Dark: unselectedPageFgDark}
}
