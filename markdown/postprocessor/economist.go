package postprocessor

import (
	"strings"
)

func processEconomist(text string) string {
	lines := strings.Split(text, "\n")
	output := ""

	for i, line := range lines {
		isOnFirstOrLastLine := i == 0 || i == len(lines)-1
		lineNoLeadingWhitespace := strings.TrimLeft(line, " ")

		if len(lineNoLeadingWhitespace) == 1 {
			continue
		}

		if strings.Contains(line, "Listen to this story") ||
			strings.Contains(line, "Your browser does not support the ") ||
			strings.Contains(line, "Listen on the go") ||
			strings.Contains(line, "Get The Economist app and play articles") ||
			strings.Contains(line, "Play in app") {
			continue
		}

		if isOnFirstOrLastLine {
			output += line + "\n"

			continue
		}

		if isOnLineBeforeTargetContains("This article appeared in the", lines, i) ||
			isOnLineBeforeTargetContains("For more coverage of ", lines, i) {
			output += "\n"

			break
		}

		output += line + "\n"
	}

	output = strings.ReplaceAll(output, "\n\n\n\n", "\n\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")
	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")

	return output
}
