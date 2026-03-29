package help

import (
	"strings"

	"charm.land/bubbles/v2/key"
)

const (
	newPar = "\n\n"
)

func HelpScreen(screenWidth int, enableNerdFonts bool, mainMenuBindings []key.Binding) string {
	var sb strings.Builder

	sb.WriteString(Text(screenWidth, enableNerdFonts, mainMenuBindings) + newPar)

	return sb.String()
}
