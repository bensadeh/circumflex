package formatter

import (
	"fmt"
	"strconv"
	"strings"
)

func GetMainText(title string, domain string) string {
	return formatTitle(title) + formatDomain(domain)
}

func formatTitle(title string) string {
	title = formatShowAndTell(title)
	title = formatYCStartups(title)
	return title
}

func formatDomain(domain string) string {
	if domain == "" {
		return ""
	}
	return dim(paren(domain))
}

func dim(text string) string {
	return "[::d]" + text + "[-:-:-]"
}

func paren(text string) string {
	return " (" + text + ")"
}

func formatShowAndTell(title string) string {
	title = strings.ReplaceAll(title, "Ask HN:", reverse("Ask HN:"))
	title = strings.ReplaceAll(title, "Show HN:", reverse("Show HN:"))
	title = strings.ReplaceAll(title, "Tell HN:", reverse("Tell HN:"))
	title = strings.ReplaceAll(title, "Launch HN:", reverse("Launch HN:"))
	return title
}

func reverse(text string) string {
	return "[::r]" + text + "[-:-:-]"
}

func formatYCStartups(title string) string {
	startYear := 05
	endYear := 21

	for i := startYear; i <= endYear; i++ {
		year := fmt.Sprintf("%02d", i)

		summer := "(YC S" + year + ")"
		winter := "(YC W" + year + ")"

		title = strings.ReplaceAll(title, summer, orange(summer))
		title = strings.ReplaceAll(title, winter, orange(winter))
	}

	return title
}

func orange(text string) string {
	return "[orange:black]" + text + "[-:-:-]"
}

func GetSecondaryText(points int, author string, time string, comments int) string {
	parsedPoints := parsePoints(points)
	parsedAuthor := parseAuthor(author)
	parsedComments := parseComments(comments, author)

	return dim(parsedPoints + parsedAuthor + time + parsedComments)
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
