package help

import (
	"clx/constants"
	"strings"
)

const (
	newPar = "\n\n"
)

func GetHelpScreen(enableNerdFonts bool) string {
	textWidth := 70

	var sb strings.Builder

	sb.WriteString(constants.InvisibleCharacterForTopLevelComments + newPar)
	sb.WriteString(constants.InvisibleCharacterForTopLevelComments + GetText(textWidth, enableNerdFonts) + newPar)

	return sb.String()
}
