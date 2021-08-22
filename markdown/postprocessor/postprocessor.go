package postprocessor

import (
	"strings"
)

func Process(text string, URL string) string {
	if strings.Contains(URL, "https://en.wikipedia.org") {
		return processWikipedia(text)
	}

	return text
}
