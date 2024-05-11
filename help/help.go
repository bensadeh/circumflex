package help

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/f01c33/circumflex/constants/unicode"
	"github.com/f01c33/circumflex/info"
)

const (
	newPar = "\n\n"
)

func GetHelpScreen(enableNerdFonts bool) string {
	textWidth := 70

	var sb strings.Builder

	sb.WriteString(unicode.InvisibleCharacterForTopLevelComments + newPar)
	sb.WriteString(unicode.InvisibleCharacterForTopLevelComments + info.GetText(textWidth, enableNerdFonts) + newPar)

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
