package postprocessor

import (
	ansi "clx/utils/strip-ansi"
	"strings"
)

func Process(text string, URL string) string {
	switch {
	case strings.Contains(URL, "en.wikipedia.org"):
		return processWikipedia(text)

	case strings.Contains(URL, "www.bbc.com"):
		return processBBC(text)

	case strings.Contains(URL, "www.nytimes.com"):
		return processNYTimes(text)

	case strings.Contains(URL, "www.economist.com"):
		return processEconomist(text)

	default:
		return text
	}
}

func isOnLineBeforeTargetEquals(target string, lines []string, i int) bool {
	nextLine := lines[i+1]
	nextLine = ansi.Strip(nextLine)
	nextLine = strings.TrimLeft(nextLine, " ")

	return nextLine == target
}

func isOnLineBeforeTargetContains(target string, lines []string, i int) bool {
	nextLine := lines[i+1]
	nextLine = ansi.Strip(nextLine)
	nextLine = strings.TrimLeft(nextLine, " ")

	return strings.Contains(nextLine, target)
}
