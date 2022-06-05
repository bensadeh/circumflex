package style

import "github.com/charmbracelet/lipgloss"

const (
	magentaDark = "200"
	yellowDark  = "214"
	blueDark    = "33"
	pinkDark    = "219"

	logoBgDark           = "#0f1429"
	labelBgDark          = logoBgDark
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
	labelBgLight          = headerBgLight
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

func GetLogoBg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: logoBgLight, Dark: logoBgDark}
}

func GetLabelBg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: labelBgLight, Dark: labelBgDark}
}

func GetHeaderBg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: headerBgLight, Dark: headerBgDark}
}

func GetStatusBarBg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: statusBarBgLight, Dark: statusBarBgDark}
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
