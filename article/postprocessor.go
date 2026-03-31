package article

import (
	"clx/layout"
	"strings"
)

func processArticle(text string, url string, width int) string {
	text = filterSite(text, url)
	text = indent(text, width)
	text = DeIndentInfoSection(text)

	return text
}

func indent(commentSection string, _ int) string {
	indentBlock := strings.Repeat(" ", layout.ReaderViewLeftMargin)

	lines := strings.Split(commentSection, "\n")
	for i, line := range lines {
		lines[i] = indentBlock + line
	}

	return strings.Join(lines, "\n")
}

// DeIndentInfoSection removes one leading space from lines containing
// info-section box-drawing characters (╭, │, ╰).
func DeIndentInfoSection(commentSection string) string {
	var sb strings.Builder

	lines := strings.Split(commentSection, "\n")

	for i, line := range lines {
		isOnLastLine := i == len(lines)-1
		isInfoSection := strings.Contains(line, "╭") || strings.Contains(line, "│") ||
			strings.Contains(line, "╰")

		if isInfoSection {
			deIndentedLine := strings.TrimPrefix(line, " ")

			sb.WriteString(deIndentedLine + "\n")

			continue
		}

		if isOnLastLine {
			continue
		}

		sb.WriteString(line + "\n")
	}

	return sb.String()
}
