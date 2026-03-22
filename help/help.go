package help

import (
	"clx/constants"
	"strings"

	"charm.land/bubbles/v2/key"
)

const (
	newPar = "\n\n"
)

const helpTextWidth = 70

func HelpScreen(enableNerdFonts bool, mainMenuBindings []key.Binding) string {
	textWidth := helpTextWidth

	var sb strings.Builder

	sb.WriteString(constants.InvisibleCharacterForTopLevelComments + newPar)
	sb.WriteString(constants.InvisibleCharacterForTopLevelComments + Text(textWidth, enableNerdFonts, mainMenuBindings) + newPar)

	return sb.String()
}
