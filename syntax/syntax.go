package syntax

import (
	"clx/colors"
	"regexp"
	"strings"

	"github.com/logrusorgru/aurora/v3"
)

const (
	askHN    = "Ask HN:"
	showHN   = "Show HN:"
	tellHN   = "Tell HN:"
	launchHN = "Launch HN:"
)

func HighlightYCStartups(comment string) string {
	expression := regexp.MustCompile(`\((YC [SW]\d{2})\)`)

	return expression.ReplaceAllString(comment, colors.OrangeBackground+colors.NearBlack+" "+`$1`+" "+colors.Normal)
}

func HighlightHackerNewsHeadlines(title string) string {
	title = strings.ReplaceAll(title, askHN, aurora.Blue(askHN).String())
	title = strings.ReplaceAll(title, showHN, aurora.Red(showHN).String())
	title = strings.ReplaceAll(title, tellHN, aurora.Magenta(tellHN).String())
	title = strings.ReplaceAll(title, launchHN, aurora.Green(launchHN).String())

	return title
}

func HighlightWhoIsHiring(title string, author string) string {
	if author != "whoishiring" {
		return title
	}

	title = strings.ReplaceAll(title, " (", colors.Normal+" ")
	title = strings.ReplaceAll(title, ")", "")

	if strings.Contains(title, "Who is hiring?") {
		title = aurora.Index(232, title).String()

		return aurora.BgBlue(title).String()
	}

	if strings.Contains(title, "Freelancer?") {
		title = aurora.Index(232, title).String()

		return aurora.BgRed(title).String()
	}

	if strings.Contains(title, "Who wants to be hired?") {
		title = aurora.Index(232, title).String()

		return aurora.BgYellow(title).String()
	}

	return title
}
