package postprocessor

import (
	"strings"
)

func processNYTimes(text string) string {
	lines := strings.Split(text, "\n")
	output := ""

	for i, line := range lines {
		isOnFirstOrLastLine := i == 0 || i == len(lines)-1
		lineNoLeadingWhitespace := strings.TrimLeft(line, " ")

		if len(lineNoLeadingWhitespace) == 1 {
			continue
		}

		if strings.Contains(line, "Credit…") ||
			lineNoLeadingWhitespace == "Credit" ||
			strings.Contains(line, "This is a developing story. Check back for updates.") {
			continue
		}

		if isOnFirstOrLastLine {
			output += line + "\n"

			continue
		}

		output += line + "\n"
	}

	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")

	return output
}
