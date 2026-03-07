package help

import (
	"clx/constants/unicode"
	"clx/info"
	"strings"
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
