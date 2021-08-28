package postprocessor

import (
	"clx/markdown/postprocessor/filter"
	"strconv"
	"strings"
)

func Process(text string, url string) string {
	switch {
	case strings.Contains(url, "en.wikipedia.org"):
		text = strings.ReplaceAll(text, "[edit]", "")
		text = removeWikipediaReferences(text)

		ruleSet := filter.RuleSet{}

		ruleSet.EndBeforeLineEquals("References")
		ruleSet.EndBeforeLineEquals("Footnotes")

		return ruleSet.Filter(text)

	case strings.Contains(url, "www.bbc.com") || strings.Contains(url, "www.bbc.co.uk"):
		return processBBC(text)

	case strings.Contains(url, "www.nytimes.com"):
		ruleSet := filter.RuleSet{}

		ruleSet.SkipLineContains("Credit…")
		ruleSet.SkipLineContains("This is a developing story. Check back for updates.")

		ruleSet.SkipLineEquals("Credit")
		ruleSet.SkipLineEquals("Image")

		return ruleSet.Filter(text)

	case strings.Contains(url, "www.economist.com"):
		ruleSet := filter.RuleSet{}

		ruleSet.SkipLineContains("Listen to this story")
		ruleSet.SkipLineContains("Your browser does not support the ")
		ruleSet.SkipLineContains("Your browser does not support the ")
		ruleSet.SkipLineContains("Listen on the go")
		ruleSet.SkipLineContains("Get The Economist app and play articles")
		ruleSet.SkipLineContains("Play in app")

		ruleSet.EndBeforeLineContains("This article appeared in the")
		ruleSet.EndBeforeLineContains("For more coverage of ")

		return ruleSet.Filter(text)

	case strings.Contains(url, "www.tomshardware.com"):
		ruleSet := filter.RuleSet{}

		ruleSet.SkipLineContains("1. Home")
		ruleSet.SkipLineContains("2. News")
		ruleSet.SkipLineContains("(Image credit: ")

		return ruleSet.Filter(text)

	case strings.Contains(url, "cnn.com"):
		ruleSet := filter.RuleSet{}

		ruleSet.SkipParContains("Credit: ")

		return ruleSet.Filter(text)

	case strings.Contains(url, "arstechnica.com"):
		ruleSet := filter.RuleSet{}

		ruleSet.SkipParContains("Enlarge/ ")

		return ruleSet.Filter(text)

	case strings.Contains(url, "macrumors.com"):
		ruleSet := filter.RuleSet{}

		ruleSet.EndBeforeLineEquals("Top Stories")

		return ruleSet.Filter(text)

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
