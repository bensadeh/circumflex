package help

import (
	"clx/constants"
	"strings"

	"charm.land/bubbles/v2/key"
)

const (
	newPar = "\n\n"
)

func GetHelpScreen(enableNerdFonts bool, mainMenuBindings []key.Binding) string {
	textWidth := 70

	var sb strings.Builder

	sb.WriteString(constants.InvisibleCharacterForTopLevelComments + newPar)
	sb.WriteString(constants.InvisibleCharacterForTopLevelComments + GetText(textWidth, enableNerdFonts, mainMenuBindings) + newPar)

	return sb.String()
}
