package article

import (
	"strconv"
	"strings"
)

func preFormatWikipediaArticle(article string) string {
	lines := strings.Split(article, newLine)
	formattedArticle := ""

	for i, line := range lines {
		isOnFirstOrLastLine := i == 0 || i == len(lines)-1

		if isOnFirstOrLastLine {
			formattedArticle += line + newLine

			continue
		}

		isBeforeReferences := isOnLineBeforeReferences(lines, i)

		if isBeforeReferences {
			break
		}

		isOnHeader := isHeader(lines, i, line)

		if isOnHeader {
			formattedArticle += strings.ReplaceAll(line, "[edit]", "") + newLine

			continue
		}

		formattedArticle += line + newLine
	}

	formattedArticle = removeReferences(formattedArticle)

	return formattedArticle
}

func isOnLineBeforeReferences(lines []string, i int) bool {
	nextLine := lines[i+1]
	nextLineLineIsReferences := nextLine == "References[edit]"

	if nextLineLineIsReferences {
		return true
	}

	return false
}

func removeReferences(input string) string {
	inputWithoutReferences := input

	for i := 1; i < 256; i++ {
		number := strconv.Itoa(i)
		inputWithoutReferences = strings.ReplaceAll(inputWithoutReferences, "^["+number+"]", "")
	}

	return inputWithoutReferences
}
