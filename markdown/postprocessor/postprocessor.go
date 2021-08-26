package postprocessor

import (
	ansi "clx/utils/strip-ansi"
	"strings"
)

func Process(text string, URL string) string {
	if strings.Contains(URL, "en.wikipedia.org") {
		return processWikipedia(text)
	}

	if strings.Contains(URL, "www.bbc.com/") {
		return processBBC(text)
	}

	return text
}

func isOnLineBeforeTarget(target string, lines []string, i int) bool {
	nextLine := lines[i+1]
	nextLine = ansi.Strip(nextLine)
	nextLine = strings.TrimLeft(nextLine, " ")
	nextLineLineIsReferences := nextLine == target

	return nextLineLineIsReferences
}
