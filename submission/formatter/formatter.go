package formatter

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	black = "#0c0c0c"
)

func GetMainText(title string, domain string, mode int) string {
	return formatTitle(title, mode) + formatDomain(domain)
}

func formatTitle(title string, mode int) string {
	switch mode {
	case 1:
		title = labelShowAndTell(title)
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

func labelShowAndTell(title string) string {
	title = strings.ReplaceAll(title, "Ask HN:", reverse("Ask HN:"))
	title = strings.ReplaceAll(title, "Show HN:", reverse("Show HN:"))
	title = strings.ReplaceAll(title, "Tell HN:", reverse("Tell HN:"))
	title = strings.ReplaceAll(title, "Launch HN:", reverse("Launch HN:"))
	return title
}

func colorizeShowAndTell(title string) string {
	title = strings.ReplaceAll(title, "Ask HN:", "[purple]"+"Ask HN:"+"[-]")
	title = strings.ReplaceAll(title, "Show HN:", "[maroon]"+"Show HN:"+"[-]")
	title = strings.ReplaceAll(title, "Tell HN:", "[navy]"+"Tell HN:"+"[-]")
	title = strings.ReplaceAll(title, "Launch HN:", "[green]"+"Launch HN:"+"[-]")
	return title
}

func reverse(text string) string {
	return "[::r]" + text + "[-:-:-]"
}

func labelYCStartups(title string) string {
	startYear := 0o5
	endYear := 22

	for i := startYear; i <= endYear; i++ {
		year := fmt.Sprintf("%02d", i)

		summer := "(YC S" + year + ")"
		winter := "(YC W" + year + ")"

		title = strings.ReplaceAll(title, summer, orange(" YC S"+year+" "))
		title = strings.ReplaceAll(title, winter, orange(" YC W"+year+" "))
	}

	return title
}

func orange(text string) string {
	return "[" + black + ":orange]" + text + "[-:-:-]"
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
