package postprocessor

import (
	"strconv"
	"strings"
)

func processWikipedia(text string) string {
	text = removeReferences(text)
	text = strings.ReplaceAll(text, "[edit]", "")

	lines := strings.Split(text, "\n")
	output := ""

	for i, line := range lines {
		isOnFirstOrLastLine := i == 0 || i == len(lines)-1

		if isOnFirstOrLastLine {
			output += line + "\n"

			continue
		}

		if isOnLineBeforeTarget("References", lines, i) || isOnLineBeforeTarget("Footnotes", lines, i) {
			output += "\n"

			break
		}

		output += line + "\n"
	}

	return output
}

func removeReferences(input string) string {
	inputWithoutReferences := input

	for i := 1; i < 256; i++ {
		number := strconv.Itoa(i)
		inputWithoutReferences = strings.ReplaceAll(inputWithoutReferences, "["+number+"]", "")
	}

	return inputWithoutReferences
}
