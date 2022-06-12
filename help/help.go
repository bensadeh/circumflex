package help

import (
	"clx/constants/unicode"
	"clx/info"
	"clx/screen"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

const (
	newLine = "\n"
	newPar  = "\n\n"
)

func GetHelpScreen() string {
	screenWidth := screen.GetTerminalWidth()
	textWidth := 70

	var sb strings.Builder

	sb.WriteString(unicode.ZeroWidthSpace + newLine + lipgloss.PlaceHorizontal(screenWidth, lipgloss.Center, "") + newPar)
	//sb.WriteString(unicode.ZeroWidthSpace + lipgloss.PlaceHorizontal(screenWidth, lipgloss.Center, getSubSection(textWidth)) + newPar)
	sb.WriteString(unicode.ZeroWidthSpace + lipgloss.PlaceHorizontal(screenWidth, lipgloss.Center, info.GetText(textWidth)) + newPar)

	return sb.String()
}

func getHeader() string {
	lg := lipgloss.NewStyle()

	return "Welcome to " + lg.
		Foreground(lipgloss.Color("5")).
		Background(lipgloss.Color("8")).
		Render(" circumflex ")
}

func getSubSection(width int) string {
	lg := lipgloss.NewStyle()

	return lg.Width(width).Align(lipgloss.Left).Background(lipgloss.Color("4")).Foreground(lipgloss.Color("16")).Render("Vivamus est arcu, porttitor ut facilisis quis, accumsan vel sem. Aenean vehicula justo a arcu porttitor posuere. Phasellus vitae pellentesque leo, in vestibulum tellus. Phasellus aliquam urna eget nisi ultrices, quis dignissim mauris blandit. Suspendisse potenti. Interdum et malesuada fames ac ante ipsum primis in faucibus. Nulla pellentesque cursus mauris, ac iaculis neque porttitor cursus. Vestibulum bibendum tempus egestas. Sed id volutpat ipsum.")
}
