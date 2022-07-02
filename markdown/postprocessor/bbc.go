package postprocessor

import (
	"strings"

	"clx/markdown/postprocessor/filter"

	. "github.com/logrusorgru/aurora/v3"
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

		if filter.IsOnLineBeforeTargetEquals([]string{"--"}, lines, i) ||
			filter.IsOnLineBeforeTargetEquals([]string{"You may also be interested in:"}, lines, i) {
			output += "\n"

			break
		}

		image := Cyan("Image: ").Faint().String()
		line = strings.ReplaceAll(line, "image source", image)

		caption := Yellow("Caption: ").Faint().String()
		line = strings.ReplaceAll(line, "image caption", caption)

		output += line + "\n"
	}

	output = strings.ReplaceAll(output, "\n\n\n", "\n\n")

	return output
}
