package formatter

import (
	"clx/constants/messages"
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

func FormatMain(title string, domain string, author string, mode int) string {
	return formatTitle(title, author, mode) + formatDomain(domain)
}

func formatTitle(title string, author string, mode int) string {
	if title == messages.EnterCommentSectionToUpdate {
		return Yellow(title)
	}

	if author == "whoishiring" {
		return highlightWhoIsHiring(title, mode)
	}

	title = strings.ReplaceAll(title, tripleSpace, singleSpace)
	title = strings.ReplaceAll(title, doubleSpace, singleSpace)
	title = strings.ReplaceAll(title, "]", "[]")

	title = highlightShowAndTell(title, mode)
	title = highlightYCStartups(title, mode)
	title = highlightSpecialContent(title, mode)

	return title
}

func highlightShowAndTell(title string, mode int) string {
	switch mode {
	case reverseHighlighting:
		title = strings.ReplaceAll(title, askHN, Reverse(askHN))
		title = strings.ReplaceAll(title, showHN, Reverse(showHN))
		title = strings.ReplaceAll(title, tellHN, Reverse(tellHN))
		title = strings.ReplaceAll(title, launchHN, Reverse(launchHN))

		return title
	case colorizedHighlighting:
		title = strings.ReplaceAll(title, askHN, Blue(askHN))
		title = strings.ReplaceAll(title, showHN, Red(showHN))
		title = strings.ReplaceAll(title, tellHN, Magenta(tellHN))
		title = strings.ReplaceAll(title, launchHN, Green(launchHN))

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
		title = strings.ReplaceAll(title, summer, Reverse(" YC S"+year+" "))
		title = strings.ReplaceAll(title, winter, Reverse(" YC W"+year+" "))
	}

	if mode == colorizedHighlighting {
		title = strings.ReplaceAll(title, summer, BlackOnOrange(" YC S"+year+" "))
		title = strings.ReplaceAll(title, winter, BlackOnOrange(" YC W"+year+" "))
	}

	return title
}

func highlightSpecialContent(title string, mode int) string {
	switch mode {
	case reverseHighlighting:
		title = strings.ReplaceAll(title, "[audio[]", Reverse("audio"))
		title = strings.ReplaceAll(title, "[video[]", Reverse("video"))
		title = strings.ReplaceAll(title, "[pdf[]", Reverse("pdf"))
		title = strings.ReplaceAll(title, "[PDF[]", Reverse("PDF"))
		title = strings.ReplaceAll(title, "[flagged[]", Reverse("flagged"))

		return title
	case colorizedHighlighting:
		title = strings.ReplaceAll(title, "[audio[]", Yellow("audio"))
		title = strings.ReplaceAll(title, "[video[]", Yellow("video"))
		title = strings.ReplaceAll(title, "[pdf[]", Yellow("pdf"))
		title = strings.ReplaceAll(title, "[PDF[]", Yellow("PDF"))
		title = strings.ReplaceAll(title, "[flagged[]", Red("flagged"))

		return title

	default:
		return title
	}
}

func highlightWhoIsHiring(title string, mode int) string {
	switch mode {
	case reverseHighlighting:
		return Reverse(title)

	case colorizedHighlighting:
		return BlackOnBlue(title)

	default:
		return title
	}
}

func formatDomain(domain string) string {
	if domain == "" {
		return ""
	}

	domainInParenthesis := " (" + domain + ")"
	domainInParenthesisAndDimmed := Dim(domainInParenthesis)

	return domainInParenthesisAndDimmed
}

func FormatSecondary(points int, author string, unixTime int64, comments int) string {
	parsedPoints := parsePoints(points)
	parsedAuthor := parseAuthor(author)
	parsedTime := parseTime(unixTime)
	parsedComments := parseComments(comments, author)

	return Dim(parsedPoints + parsedAuthor + parsedTime + parsedComments)
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
