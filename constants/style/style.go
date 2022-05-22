package style

import "github.com/charmbracelet/lipgloss"

const (
	MagentaDark = "200"
	YellowDark  = "214"
	BlueDark    = "33"
	PinkDark    = "219"

	LogoBackgroundDark   = "234"
	HeaderBackgroundDark = "235"
	UnselectedItemDark   = "250"
	SelectedPageDark     = UnselectedItemDark
	UnselectedPageDark   = "239"

	MagentaLight = MagentaDark
	YellowLight  = "208"
	BlueLight    = BlueDark
	PinkLight    = PinkDark

	LogoBackgroundLight   = "254"
	HeaderBackgroundLight = "247"
	UnselectedItemLight   = "235"
	SelectedPageLight     = UnselectedItemLight
	UnselectedPageLight   = "247"
)

func GetMagenta() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: MagentaLight, Dark: MagentaDark}
}

func GetYellow() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: YellowLight, Dark: YellowDark}
}

func GetBlue() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: BlueLight, Dark: BlueDark}
}

func GetPink() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: PinkLight, Dark: PinkDark}
}

func GetLogoBackground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: LogoBackgroundLight, Dark: LogoBackgroundDark}
}

func GetHeaderBackground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: HeaderBackgroundLight, Dark: HeaderBackgroundDark}
}

func GetUnselectedItemForeground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: UnselectedItemLight, Dark: UnselectedItemDark}
}

func GetSelectedPageForeground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: SelectedPageLight, Dark: SelectedPageDark}
}

func GetUnselectedPageForeground() lipgloss.TerminalColor {
	return lipgloss.AdaptiveColor{Light: UnselectedPageLight, Dark: UnselectedPageDark}
}
