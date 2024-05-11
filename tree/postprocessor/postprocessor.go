package postprocessor

import (
	"strings"

	"github.com/f01c33/clx/constants/margins"
	"github.com/f01c33/clx/constants/unicode"

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

	return strings.ReplaceAll(commentSection, newLine+indentBlock+unicode.InvisibleCharacterForTopLevelComments,
		unicode.InvisibleCharacterForTopLevelComments+newLine+indentBlock)
}

func indentCommentSection(commentSection string, screenWidth int) string {
	indentBlock := getIndentBlock()

	indentedCommentSection, _ := text.WrapWithPad(commentSection, screenWidth, indentBlock)

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

func getIndentBlock() string {
	return strings.Repeat(" ", margins.CommentSectionLeftMargin)
}
