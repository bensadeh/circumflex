package postprocessor

import (
	"clx/constants/margins"
	"clx/constants/unicode"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

const (
	newLine = "\n"
)

func Process(commentSection string, screenWidth int) string {
	commentSection = indentCommentSection(commentSection, screenWidth)
	commentSection = moveZeroWidthSpaceUpOneLine(commentSection)
	commentSection = deIndentInfoSection(commentSection)

	return commentSection
}

func moveZeroWidthSpaceUpOneLine(commentSection string) string {
	indentBlock := getIndentBlock()

	return strings.ReplaceAll(commentSection, newLine+indentBlock+unicode.ZeroWidthSpace,
		unicode.ZeroWidthSpace+newLine+indentBlock)
}

func indentCommentSection(commentSection string, screenWidth int) string {
	indentBlock := getIndentBlock()

	indentedCommentSection, _ := text.WrapWithPad(commentSection, screenWidth, indentBlock)

	return indentedCommentSection
}

func deIndentInfoSection(commentSection string) string {
	var sb strings.Builder

	lines := strings.Split(commentSection, "\n")

	for _, line := range lines {
		isInfoSection := strings.Contains(line, "╭") || strings.Contains(line, "│") ||
			strings.Contains(line, "╰")

		if isInfoSection {
			deIndentedLine := strings.TrimPrefix(line, " ")

			sb.WriteString(deIndentedLine + "\n")

			continue
		}

		sb.WriteString(line + "\n")
	}

	return sb.String()
}

func getIndentBlock() string {
	return strings.Repeat(" ", margins.CommentSectionLeftMargin)
}
