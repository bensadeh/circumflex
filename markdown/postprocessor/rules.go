package postprocessor

import (
	"strings"

	"clx/markdown/postprocessor/filter"
)

func filterSite(text string, url string) string {
	ruleSet := filter.RuleSet{}

	switch {
	case strings.Contains(url, "en.wikipedia.org"):
		text = strings.ReplaceAll(text, "[edit]", "")
		text = removeWikipediaReferences(text)

		ruleSet.EndBeforeLineEquals("References")
		ruleSet.EndBeforeLineEquals("Footnotes")

		return ruleSet.Filter(text)

	case strings.Contains(url, "bbc.com") || strings.Contains(url, "bbc.co.uk"):
		return processBBC(text)

	case strings.Contains(url, "nytimes.com"):
		ruleSet.SkipParContains("Credit…")
		ruleSet.SkipParContains("This is a developing story. Check back for updates.")

		ruleSet.SkipLineEquals("Credit")
		ruleSet.SkipLineEquals("Image")

		return ruleSet.Filter(text)

	case strings.Contains(url, "economist.com"):
		ruleSet.SkipParContains("Listen to this story")
		ruleSet.SkipParContains("Your browser does not support the ")
		ruleSet.SkipParContains("Listen on the go")
		ruleSet.SkipParContains("Get The Economist app and play articles")
		ruleSet.SkipParContains("Play in app")
		ruleSet.SkipParContains("Enjoy more audio and podcasts on iOS or Android")

		ruleSet.EndBeforeLineContains("This article appeared in the")
		ruleSet.EndBeforeLineContains("For more coverage of ")

		return ruleSet.Filter(text)

	case strings.Contains(url, "tomshardware.com"):
		ruleSet.SkipParContains("1. Home")
		ruleSet.SkipParContains("2. News")
		ruleSet.SkipParContains("(Image credit: ")

		return ruleSet.Filter(text)

	case strings.Contains(url, "cnn.com"):
		ruleSet.SkipParContains("Credit: ")

		return ruleSet.Filter(text)

	case strings.Contains(url, "arstechnica.com"):
		ruleSet.SkipParContains("Enlarge/ ")
		ruleSet.SkipParContains("This story originally appeared on ")

		return ruleSet.Filter(text)

	case strings.Contains(url, "macrumors.com"):
		ruleSet.EndBeforeLineEquals("Top Stories")
		ruleSet.EndBeforeLineEquals("Related Stories")

		return ruleSet.Filter(text)

	case strings.Contains(url, "wired.com") || strings.Contains(url, "wired.co.uk"):
		ruleSet.SkipParContains("Read more: ")
		ruleSet.SkipParContains("Do you use social media regularly? Take our short survey.")

		ruleSet.EndBeforeLineEquals("More Great WIRED Stories")

		return ruleSet.Filter(text)

	case strings.Contains(url, "theguardian.com"):
		ruleSet.SkipParContains("Photograph:")

		return ruleSet.Filter(text)

	case strings.Contains(url, "axios.com"):
		ruleSet.SkipParContains("Sign up for our daily briefing")
		ruleSet.SkipParContains("Catch up on the day's biggest business stories")
		ruleSet.SkipParContains("Stay on top of the latest market trends")
		ruleSet.SkipParContains("Sports news worthy of your time")
		ruleSet.SkipParContains("Tech news worthy of your time")
		ruleSet.SkipParContains("Get the inside stories")
		ruleSet.SkipParContains("Axios on your phone")
		ruleSet.SkipParContains("Catch up on coronavirus stories and special reports")
		ruleSet.SkipParContains("Want a daily digest of the top ")
		ruleSet.SkipParContains("Get a daily digest of the most important stories ")
		ruleSet.SkipParContains("Download for free.")
		ruleSet.SkipParContains("Sign up for free.")
		ruleSet.SkipParContains("Make your busy days simpler with Axios AM/PM")
		ruleSet.SkipParContains("Subscribe to Axios Closer")
		ruleSet.SkipParContains("Get breaking news")
		ruleSet.SkipParContains("Sign up for Axios")
		ruleSet.SkipParContains("Stay up-to-date on the most important and interesting")

		return ruleSet.Filter(text)

	case strings.Contains(url, "9to5mac.com"):
		ruleSet.SkipParContains("We use income earning auto affiliate links.")
		ruleSet.SkipParContains("Check out 9to5Mac on YouTube for more Apple news:")

		ruleSet.EndBeforeLineEquals("About the Author")

		return ruleSet.Filter(text)

	case strings.Contains(url, "smithsonianmag.com"):
		ruleSet.SkipParContains("smithsonianmag.com")

		ruleSet.EndBeforeLineEquals("Like this article?")

		return ruleSet.Filter(text)

	case strings.Contains(url, "cnet.com"):
		ruleSet.SkipParContains("Read more:")
		ruleSet.SkipParContains("Stay up-to-date on the latest news")

		return ruleSet.Filter(text)

	default:
		return text
	}
}
