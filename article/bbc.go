package article

import (
	"clx/style"
	"strings"

	"charm.land/lipgloss/v2"
)

func processBBC(text string) string {
	lines := strings.Split(text, "\n")

	var sb strings.Builder

	for i, line := range lines {
		isOnFirstOrLastLine := i == 0 || i == len(lines)-1
		lineNoLeadingWhitespace := strings.TrimLeft(line, " ")

		if len(lineNoLeadingWhitespace) == 1 {
			continue
		}

		if strings.Contains(line, "(Image credit: ") {
			continue
		}

		if isOnFirstOrLastLine {
			sb.WriteString(line)
			sb.WriteByte('\n')

			continue
		}

		if isOnLineBeforeTargetEquals([]string{"--"}, lines, i) ||
			isOnLineBeforeTargetEquals([]string{"You may also be interested in:"}, lines, i) {
			sb.WriteByte('\n')

			break
		}

		image := lipgloss.NewStyle().Foreground(style.ReaderBBCImageColor()).Faint(true).Render("Image: ")
		line = strings.ReplaceAll(line, "image source", image)

		caption := lipgloss.NewStyle().Foreground(style.ReaderBBCCaptionColor()).Faint(true).Render("Caption: ")
		line = strings.ReplaceAll(line, "image caption", caption)

		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	output := strings.ReplaceAll(sb.String(), "\n\n\n", "\n\n")

	return output
}
