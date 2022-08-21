package postprocessor

import (
	"strconv"
	"strings"
)

func removeWikipediaReferences(input string) string {
	inputWithoutReferences := input

	for i := 1; i < 256; i++ {
		number := strconv.Itoa(i)
		inputWithoutReferences = strings.ReplaceAll(inputWithoutReferences, "\\["+number+"\\]", "")
	}

	return inputWithoutReferences
}
