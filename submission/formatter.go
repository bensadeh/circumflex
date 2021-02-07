package submission

import (
	"clx/utils/format"
	"fmt"
	"strconv"
	"strings"
)

const (
	askHN    = "Ask HN:"
	showHN   = "Show HN:"
	tellHN   = "Tell HN:"
	launchHN = "Launch HN:"
)

func FormatSubMain(title string, domain string, mode int) string {
	return formatTitle(title, mode) + formatDomain(domain)
}

func formatTitle(title string, mode int) string {
	switch mode {
	case 1:
		title = reverseShowAndTell(title)
		title = labelYCStartups(title)

		return title
	case 2:
		title = colorizeShowAndTell(title)
		title = labelYCStartups(title)

		return title
	default:
		return title
	}
}

func reverseShowAndTell(title string) string {
	title = strings.ReplaceAll(title, askHN, format.Reverse(askHN))
	title = strings.ReplaceAll(title, showHN, format.Reverse(showHN))
	title = strings.ReplaceAll(title, tellHN, format.Reverse(tellHN))
	title = strings.ReplaceAll(title, launchHN, format.Reverse(launchHN))

	return title
}

func colorizeShowAndTell(title string) string {
	title = strings.ReplaceAll(title, askHN, format.Magenta(askHN))
	title = strings.ReplaceAll(title, showHN, format.Red(showHN))
	title = strings.ReplaceAll(title, tellHN, format.Blue(tellHN))
	title = strings.ReplaceAll(title, launchHN, format.Green(launchHN))

	return title
}

func labelYCStartups(title string) string {
	startYear, endYear := 0o5, 22

	for i := startYear; i <= endYear; i++ {
		year := fmt.Sprintf("%02d", i)

		summer := "(YC S" + year + ")"
		winter := "(YC W" + year + ")"

		title = strings.ReplaceAll(title, summer, format.BlackOnOrange(" YC S"+year+" "))
		title = strings.ReplaceAll(title, winter, format.BlackOnOrange(" YC W"+year+" "))
	}

	return title
}

func formatDomain(domain string) string {
	if domain == "" {
		return ""
	}

	domainInParenthesis := " (" + domain + ")"
	domainInParenthesisAndDimmed := format.Dim(domainInParenthesis)

	return domainInParenthesisAndDimmed
}

func FormatSubSecondary(points int, author string, time string, comments int) string {
	parsedPoints := parsePoints(points)
	parsedAuthor := parseAuthor(author)
	parsedComments := parseComments(comments, author)

	return format.Dim(parsedPoints + parsedAuthor + time + parsedComments)
}

func parseComments(comments int, author string) string {
	if author == "" {
		return ""
	}

	c := strconv.Itoa(comments)

	return " | " + c + " comments"
}

func parseAuthor(author string) string {
	if author == "" {
		return ""
	}

	return "by " + author + " "
}

func parsePoints(points int) string {
	if points == 0 {
		return ""
	}

	return strconv.Itoa(points) + " points "
}
