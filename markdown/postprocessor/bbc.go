package postprocessor

import (
	"strings"
)

func processBBC(text string) string {
	lines := strings.Split(text, "\n")
	output := ""

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
			output += line + "\n"

			continue
		}

		if isOnLineBeforeTarget("--", lines, i) {
			output += "\n"

			break
		}

		output += line + "\n"
	}

	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")

	return output
}
