package style

import "github.com/charmbracelet/lipgloss"

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
	statusBarBgDark         = headerBgDark
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
	statusBarBgLight         = headerBgLight
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

func GetLabelBg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: labelBgLight, Dark: labelBgDark}
}

func GetLabelFg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: labelFgLight, Dark: labelFgDark}
}

func GetLabelMarkAsReadFg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: labelMarkAsReadFgLight, Dark: labelMarkAsReadFgDark}
}

func GetLabelMarkAsReadBg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: labelMarkAsReadBgLight, Dark: labelMarkAsReadBgDark}
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

func GetYCLogoFg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: ycLogoFgLight, Dark: ycLogoFgDark}
}

func GetYCLogoMarkAsReadFg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: ycLogoMarkAsReadFgLight, Dark: ycLogoMarkAsReadFgDark}
}

func GetYCLabelBg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: ycLabelBgLight, Dark: ycLabelBgDark}
}

func GetYCLabelMarkAsReadBg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: ycLabelMarkAsReadBgLight, Dark: ycLabelMarkAsReadBgDark}
}

func GetYCTextFg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: ycTextFgLight, Dark: ycTextFgDark}
}

func GetYCTextMarkAsReadFg() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: ycTextMarkAsReadFgLight, Dark: ycTextMarkAsReadFgDark}
}
