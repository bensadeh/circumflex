package formatter

import (
	"clx/constants/messages"
	"fmt"
	"strconv"
	"strings"

	"github.com/nleeper/goment"
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

func FormatMain(title string, domain string, author string, highlightHeadlines bool, markAsRead bool) string {
	readModifier := ""

	if markAsRead {
		readModifier = "[::di]"
	}

	return readModifier + formatTitle(title, author, highlightHeadlines) + formatDomain(domain, markAsRead)
}

func formatTitle(title string, author string, highlightHeadlines bool) string {
	if title == messages.EnterCommentSectionToUpdate {
		return Yellow(title)
	}

	if author == "whoishiring" {
		return highlightWhoIsHiring(title)
	}

	title = strings.ReplaceAll(title, tripleSpace, singleSpace)
	title = strings.ReplaceAll(title, doubleSpace, singleSpace)
	title = strings.ReplaceAll(title, "]", "[]")

	if highlightHeadlines {
		title = highlightShowAndTell(title)
		title = highlightYCStartups(title)
		title = highlightSpecialContent(title)
	}

	return title
}

func highlightShowAndTell(title string) string {
	title = strings.ReplaceAll(title, askHN, Blue(askHN))
	title = strings.ReplaceAll(title, showHN, Red(showHN))
	title = strings.ReplaceAll(title, tellHN, Magenta(tellHN))
	title = strings.ReplaceAll(title, launchHN, Green(launchHN))

	return title
}

func highlightYCStartups(title string) string {
	startYear, endYear := 0o5, 22

	for i := startYear; i <= endYear; i++ {
		year := fmt.Sprintf("%02d", i)

		summer := "(YC S" + year + ")"
		winter := "(YC W" + year + ")"

		title = strings.ReplaceAll(title, summer, BlackOnOrange(" YC S"+year+" "))
		title = strings.ReplaceAll(title, winter, BlackOnOrange(" YC W"+year+" "))
	}

	return title
}

func highlightSpecialContent(title string) string {
	title = strings.ReplaceAll(title, "[audio[]", Yellow("audio"))
	title = strings.ReplaceAll(title, "[video[]", Yellow("video"))
	title = strings.ReplaceAll(title, "[pdf[]", Yellow("pdf"))
	title = strings.ReplaceAll(title, "[PDF[]", Yellow("PDF"))
	title = strings.ReplaceAll(title, "[flagged[]", Red("flagged"))

	return title
}

func highlightWhoIsHiring(title string) string {
	title = strings.ReplaceAll(title, " (", "[-:-:] (")

	if strings.Contains(title, "Who is hiring?") {
		return BlackOnBlue(title)
	}

	if strings.Contains(title, "Freelancer?") {
		return BlackOnRed(title)
	}

	if strings.Contains(title, "Who wants to be hired?") {
		return BlackOnYellow(title)
	}

	return title
}

func formatDomain(domain string, markAsRead bool) string {
	if domain == "" {
		return ""
	}

	readModifier := ""

	if markAsRead {
		readModifier = "[::di]"
	}

	domainInParenthesis := " (" + domain + ")"
	domainInParenthesisAndDimmed := readModifier + Dim(readModifier+domainInParenthesis)

	return domainInParenthesisAndDimmed
}

func FormatSecondary(points int, author string, unixTime int64, comments int, highlightHeadlines bool) string {
	parsedPoints := parsePoints(points)
	parsedAuthor := parseAuthor(author, highlightHeadlines)
	parsedTime := parseTime(unixTime)
	parsedComments := parseComments(comments, author)

	return "[::d]" + parsedPoints + parsedAuthor + parsedTime + parsedComments
}

func parsePoints(points int) string {
	if points == 0 {
		return ""
	}

	return strconv.Itoa(points) + " points "
}

func parseAuthor(author string, highlightHeadlines bool) string {
	if author == "" {
		return ""
	}

	if highlightHeadlines && author == "dang" {
		return "by " + Green(author) + " "
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
