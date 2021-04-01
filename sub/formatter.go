package sub

import (
	"clx/utils/format"
	"fmt"
	"strconv"
	"strings"

	"github.com/nleeper/goment"
)

const (
	noHighlighting        = 0
	reverseHighlighting   = 1
	colorizedHighlighting = 2
	askHN                 = "Ask HN:"
	showHN                = "Show HN:"
	tellHN                = "Tell HN:"
	launchHN              = "Launch HN:"
	tripleSpace           = "   "
	doubleSpace           = "  "
	singleSpace           = " "
)

func FormatSubMain(title string, domain string, mode int) string {
	return formatTitle(title, mode) + formatDomain(domain)
}

func formatTitle(title string, mode int) string {
	title = strings.ReplaceAll(title, tripleSpace, singleSpace)
	title = strings.ReplaceAll(title, doubleSpace, singleSpace)
	title = strings.ReplaceAll(title, "]", "[]")

	title = highlightShowAndTell(title, mode)
	title = highlightYCStartups(title, mode)

	return title
}

func highlightShowAndTell(title string, mode int) string {
	switch mode {
	case reverseHighlighting:
		title = strings.ReplaceAll(title, askHN, format.Reverse(askHN))
		title = strings.ReplaceAll(title, showHN, format.Reverse(showHN))
		title = strings.ReplaceAll(title, tellHN, format.Reverse(tellHN))
		title = strings.ReplaceAll(title, launchHN, format.Reverse(launchHN))

		return title
	case colorizedHighlighting:
		title = strings.ReplaceAll(title, askHN, format.Magenta(askHN))
		title = strings.ReplaceAll(title, showHN, format.Red(showHN))
		title = strings.ReplaceAll(title, tellHN, format.Blue(tellHN))
		title = strings.ReplaceAll(title, launchHN, format.Green(launchHN))

		return title

	default:
		return title
	}
}

func highlightYCStartups(title string, mode int) string {
	if mode == noHighlighting {
		return title
	}

	startYear, endYear := 0o5, 22

	for i := startYear; i <= endYear; i++ {
		year := fmt.Sprintf("%02d", i)

		summer := "(YC S" + year + ")"
		winter := "(YC W" + year + ")"

		title = formatStartup(title, mode, summer, year, winter)
	}

	return title
}

func formatStartup(title string, mode int, summer string, year string, winter string) string {
	if mode == reverseHighlighting {
		title = strings.ReplaceAll(title, summer, format.Reverse(" YC S"+year+" "))
		title = strings.ReplaceAll(title, winter, format.Reverse(" YC W"+year+" "))
	}

	if mode == colorizedHighlighting {
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

func FormatSubSecondary(points int, author string, unixTime int64, comments int) string {
	parsedPoints := parsePoints(points)
	parsedAuthor := parseAuthor(author)
	parsedTime := parseTime(unixTime)
	parsedComments := parseComments(comments, author)

	return format.Dim(parsedPoints + parsedAuthor + parsedTime + parsedComments)
}

func parsePoints(points int) string {
	if points == 0 {
		return ""
	}

	return strconv.Itoa(points) + " points "
}

func parseAuthor(author string) string {
	if author == "" {
		return ""
	}

	return "by " + author + " "
}

func parseTime(unixTime int64) string {
	moment, _ := goment.Unix(unixTime)
	now, _ := goment.New()

	return moment.From(now)
}

func parseComments(comments int, author string) string {
	if author == "" {
		return ""
	}

	c := strconv.Itoa(comments)

	return " | " + c + " comments"
}
