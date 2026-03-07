package reader

import (
	"clx/constants"
	"strings"
)

func filterSite(text string, url string) string {
	rs := ruleSet{}

	switch {
	case strings.Contains(url, "en.wikipedia.org"):
		text = strings.ReplaceAll(text, "[edit]", "")
		text = removeWikipediaReferences(text)

		rs.endBeforeLineEquals(constants.Block + " References")
		rs.endBeforeLineEquals(constants.Block + " Footnotes")
		rs.endBeforeLineEquals(constants.Block + " See also")
		rs.endBeforeLineEquals(constants.Block + " Notes")

		return rs.filter(text)

	case strings.Contains(url, "bbc.com") || strings.Contains(url, "bbc.co.uk"):
		return processBBC(text)

	case strings.Contains(url, "nytimes.com"):
		rs.skipParContains("Credit…")
		rs.skipParContains("This is a developing story. Check back for updates.")

		rs.skipLineEquals("Credit")
		rs.skipLineEquals("Image")

		return rs.filter(text)

	case strings.Contains(url, "economist.com"):
		rs.skipParContains("Listen to this story")
		rs.skipParContains("Your browser does not support the ")
		rs.skipParContains("Listen on the go")
		rs.skipParContains("Get The Economist app and play articles")
		rs.skipParContains("Play in app")
		rs.skipParContains("Enjoy more audio and podcasts on iOS or Android")

		rs.endBeforeLineContains("This article appeared in the")
		rs.endBeforeLineContains("For more coverage of ")

		return rs.filter(text)

	case strings.Contains(url, "tomshardware.com"):
		rs.skipParContains("1. Home")
		rs.skipParContains("2. News")
		rs.skipParContains("(Image credit: ")

		return rs.filter(text)

	case strings.Contains(url, "cnn.com"):
		rs.skipParContains("Credit: ")

		return rs.filter(text)

	case strings.Contains(url, "arstechnica.com"):
		rs.skipParContains("Enlarge/ ")
		rs.skipParContains("This story originally appeared on ")

		return rs.filter(text)

	case strings.Contains(url, "macrumors.com"):
		rs.endBeforeLineEquals("Top Stories")
		rs.endBeforeLineEquals("Related Stories")

		return rs.filter(text)

	case strings.Contains(url, "wired.com") || strings.Contains(url, "wired.co.uk"):
		rs.skipParContains("Read more: ")
		rs.skipParContains("Do you use social media regularly? Take our short survey.")

		rs.endBeforeLineEquals("More Great WIRED Stories")

		return rs.filter(text)

	case strings.Contains(url, "theguardian.com"):
		rs.skipParContains("Photograph:")

		return rs.filter(text)

	case strings.Contains(url, "axios.com"):
		rs.skipParContains("Sign up for our daily briefing")
		rs.skipParContains("Catch up on the day's biggest business stories")
		rs.skipParContains("Stay on top of the latest market trends")
		rs.skipParContains("Sports news worthy of your time")
		rs.skipParContains("Tech news worthy of your time")
		rs.skipParContains("Get the inside stories")
		rs.skipParContains("Axios on your phone")
		rs.skipParContains("Catch up on coronavirus stories and special reports")
		rs.skipParContains("Want a daily digest of the top ")
		rs.skipParContains("Get a daily digest of the most important stories ")
		rs.skipParContains("Download for free.")
		rs.skipParContains("Sign up for free.")
		rs.skipParContains("Make your busy days simpler with Axios AM/PM")
		rs.skipParContains("Subscribe to Axios Closer")
		rs.skipParContains("Get breaking news")
		rs.skipParContains("Sign up for Axios")
		rs.skipParContains("Stay up-to-date on the most important and interesting")

		return rs.filter(text)

	case strings.Contains(url, "9to5mac.com"):
		rs.skipParContains("We use income earning auto affiliate links.")
		rs.skipParContains("Check out 9to5Mac on YouTube for more Apple news:")

		rs.endBeforeLineEquals("About the Author")

		return rs.filter(text)

	case strings.Contains(url, "smithsonianmag.com"):
		rs.skipParContains("smithsonianmag.com")

		rs.endBeforeLineEquals("Like this article?")

		return rs.filter(text)

	case strings.Contains(url, "cnet.com"):
		rs.skipParContains("Read more:")
		rs.skipParContains("Stay up-to-date on the latest news")

		return rs.filter(text)

	default:
		return text
	}
}
