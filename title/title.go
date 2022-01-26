package formatter

import (
	"clx/constants/messages"
	"clx/utils/formatter"
	"regexp"
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

func FormatMain(title string, domain string, highlightHeadlines bool, markAsRead bool) string {
	readModifier := ""

	if markAsRead {
		readModifier = "[::di]"
	}

	return readModifier + formatTitle(title, highlightHeadlines) + formatDomain(domain, markAsRead)
}

func formatTitle(title string, highlightHeadlines bool) string {
	if title == messages.EnterCommentSectionToUpdate {
		return formatter.Yellow(title)
	}

	title = strings.ReplaceAll(title, tripleSpace, singleSpace)
	title = strings.ReplaceAll(title, doubleSpace, singleSpace)
	title = strings.ReplaceAll(title, "]", "[]")

	if highlightHeadlines {
		title = highlightShowAndTell(title)
		title = highlightYCStartups(title)
		title = highlightYear(title)
		title = highlightSpecialContent(title)
	}

	return title
}

func highlightShowAndTell(title string) string {
	title = strings.ReplaceAll(title, askHN, formatter.Blue(askHN))
	title = strings.ReplaceAll(title, showHN, formatter.Red(showHN))
	title = strings.ReplaceAll(title, tellHN, formatter.Magenta(tellHN))
	title = strings.ReplaceAll(title, launchHN, formatter.Green(launchHN))

	return title
}

func highlightYCStartups(title string) string {
	expression := regexp.MustCompile(`\((YC [SW]\d{2})\)`)

	firstHighlightGroup := `$1`
	highlightedStartup := formatter.BlackOnOrange(" " + firstHighlightGroup + " ")

	return expression.ReplaceAllString(title, highlightedStartup)
}

func highlightYear(title string) string {
	expression := regexp.MustCompile(`\((\d{4})\)`)

	firstHighlightGroup := `$1`
	highlightedYear := formatter.Year(" " + firstHighlightGroup + " ")

	return expression.ReplaceAllString(title, highlightedYear)
}

func highlightSpecialContent(title string) string {
	title = strings.ReplaceAll(title, "[audio[]", formatter.Cyan("audio"))
	title = strings.ReplaceAll(title, "[video[]", formatter.Cyan("video"))
	title = strings.ReplaceAll(title, "[pdf[]", formatter.Cyan("pdf"))
	title = strings.ReplaceAll(title, "[PDF[]", formatter.Cyan("PDF"))
	title = strings.ReplaceAll(title, "[flagged[]", formatter.Red("flagged"))

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
	domainInParenthesisAndDimmed := readModifier + formatter.Dim(readModifier+domainInParenthesis)

	return domainInParenthesisAndDimmed
}

func FormatSecondary(points int, author string, unixTime int64, comments int, commentsOld int, highlightHeadlines bool) string {
	parsedPoints := parsePoints(points)
	parsedAuthor := parseAuthor(author, highlightHeadlines)
	parsedTime := parseTime(unixTime)
	parsedComments := parseComments(comments, commentsOld, author)

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
		return "by " + formatter.Green(author) + " "
	}

	return "by " + author + " "
}

func parseTime(unixTime int64) string {
	moment, _ := goment.Unix(unixTime)
	now, _ := goment.New()

	return moment.From(now)
}

func parseComments(comments int, commentsOld int, author string) string {
	if author == "" {
		return ""
	}

	c := strconv.Itoa(comments)

	commentsDiff := comments - commentsOld

	if commentsDiff > 0 && commentsOld != 0 {
		return " | " + c + " comments • [yellow]" + strconv.Itoa(commentsDiff) + " new"
	}

	return " | " + c + " comments"
}
