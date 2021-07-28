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

	firstHighlightGroup := `$1`
	highlightedStartup := colors.OrangeBackground + colors.NearBlack + " " + firstHighlightGroup + " " + colors.Normal

	return expression.ReplaceAllString(comment, highlightedStartup)
}

func HighlightHackerNewsHeadlines(title string) string {
	title = strings.ReplaceAll(title, askHN, aurora.Blue(askHN).String())
	title = strings.ReplaceAll(title, showHN, aurora.Red(showHN).String())
	title = strings.ReplaceAll(title, tellHN, aurora.Magenta(tellHN).String())
	title = strings.ReplaceAll(title, launchHN, aurora.Green(launchHN).String())

	return title
}

func HighlightSpecialContent(title string) string {
	title = strings.ReplaceAll(title, "[audio]", aurora.Yellow("audio").String())
	title = strings.ReplaceAll(title, "[video]", aurora.Yellow("video").String())
	title = strings.ReplaceAll(title, "[pdf]", aurora.Yellow("pdf").String())
	title = strings.ReplaceAll(title, "[PDF]", aurora.Yellow("PDF").String())
	title = strings.ReplaceAll(title, "[flagged]", aurora.Red("flagged").String())

	return title
}

func HighlightWhoIsHiring(title string, author string) string {
	if author != "whoishiring" {
		return title
	}

	title = strings.ReplaceAll(title, " (", colors.Normal+" (")

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
