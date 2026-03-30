package help

import (
	"strings"

	"charm.land/bubbles/v2/key"
)

const (
	newPar = "\n\n"
)

// FitToHeight pads or truncates content to exactly height lines.
// The returned string contains height lines joined by \n with no trailing newline.
func FitToHeight(content string, height int) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")

	if len(lines) > height {
		lines = lines[:height]
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func MainMenuHelpScreen(screenWidth int, mainMenuBindings []key.Binding) string {
	var sb strings.Builder

	sb.WriteString(MainMenuText(screenWidth, mainMenuBindings) + newPar)

	return sb.String()
}

func ReaderHelpScreen(screenWidth int) string {
	var sb strings.Builder

	sb.WriteString(ReaderText(screenWidth) + newPar)

	return sb.String()
}

func CommentHelpScreen(screenWidth int, enableNerdFonts bool) string {
	var sb strings.Builder

	sb.WriteString(CommentText(screenWidth, enableNerdFonts) + newPar)

	return sb.String()
}
