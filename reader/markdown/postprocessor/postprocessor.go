package postprocessor

import (
	"strings"

	"clx/constants/margins"
	"clx/constants/unicode"
	"clx/screen"

	t "github.com/MichaelMure/go-term-text"
)

const (
	newLine = "\n"
)

func Process(text string, url string) string {
	text = filterSite(text, url)
	text = moveZeroWidthSpaceUpOneLine(text)
	text = indent(text)
	text = deIndentInfoSection(text)

	return text
}

func moveZeroWidthSpaceUpOneLine(text string) string {
	return strings.ReplaceAll(text, newLine+unicode.ZeroWidthSpace,
		unicode.ZeroWidthSpace+newLine)
}

func indent(commentSection string) string {
	indentBlock := strings.Repeat(" ", margins.ReaderViewLeftMargin)
	screenWidth := screen.GetTerminalWidth()

	indentedCommentSection, _ := t.WrapWithPad(commentSection, screenWidth, indentBlock)

	return indentedCommentSection
}

func deIndentInfoSection(commentSection string) string {
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
