package syntax

import (
	"clx/colors"
	"regexp"
	"strings"

	"github.com/logrusorgru/aurora/v3"
)

const (
	askHN       = "Ask HN:"
	showHN      = "Show HN:"
	tellHN      = "Tell HN:"
	launchHN    = "Launch HN:"
	tripleSpace = "   "
	doubleSpace = "  "
	singleSpace = " "
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
