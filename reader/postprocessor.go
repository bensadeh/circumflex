package reader

import (
	"clx/constants"
	"strings"

	t "github.com/MichaelMure/go-term-text"
)

func processArticle(text string, url string, width int) string {
	text = filterSite(text, url)
	text = indent(text, width)
	text = DeIndentInfoSection(text)

	return text
}

func indent(commentSection string, contentWidth int) string {
	indentBlock := strings.Repeat(" ", constants.ReaderViewLeftMargin)
	indentedCommentSection, _ := t.WrapWithPad(commentSection, contentWidth+constants.ReaderViewLeftMargin, indentBlock)

	return indentedCommentSection
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
