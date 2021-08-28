package postprocessor

import (
	"strconv"
	"strings"
)

func Process(text string, URL string) string {
	switch {
	case strings.Contains(URL, "en.wikipedia.org"):
		text = strings.ReplaceAll(text, "[edit]", "")
		text = removeWikipediaReferences(text)

		rs := ruleSet{
			skipContains: nil,
			skipEquals:   nil,
			endContains:  nil,
			endEquals:    []string{"References", "Footnotes"},
		}

		return rs.process(text)

	case strings.Contains(URL, "www.bbc.com") || strings.Contains(URL, "www.bbc.co.uk"):
		return processBBC(text)

	case strings.Contains(URL, "www.nytimes.com"):
		rs := ruleSet{
			skipContains: []string{"Credit…", "This is a developing story. Check back for updates."},
			skipEquals:   []string{"Credit"},
			endContains:  nil,
			endEquals:    nil,
		}

		return rs.process(text)

	case strings.Contains(URL, "www.economist.com"):
		rs := ruleSet{
			skipContains: []string{
				"Listen to this story", "Your browser does not support the ", "Listen on the go",
				"Get The Economist app and play articles", "Play in app",
			},
			skipEquals:  nil,
			endContains: []string{"This article appeared in the", "For more coverage of "},
			endEquals:   nil,
		}

		return rs.process(text)

	case strings.Contains(URL, "www.tomshardware.com"):
		rs := ruleSet{
			skipContains: []string{"(Image credit: "},
			skipEquals:   nil,
			endContains:  nil,
			endEquals:    nil,
		}

		return rs.process(text)

	default:
		return text
	}
}

func removeWikipediaReferences(input string) string {
	inputWithoutReferences := input

	for i := 1; i < 256; i++ {
		number := strconv.Itoa(i)
		inputWithoutReferences = strings.ReplaceAll(inputWithoutReferences, "["+number+"]", "")
	}

	return inputWithoutReferences
}
