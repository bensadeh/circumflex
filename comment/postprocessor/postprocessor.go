package postprocessor

import (
	"clx/constants/margins"
	"clx/constants/unicode"
	"strings"
)

const (
	newLine = "\n"
)

func Process(commentSection string) string {
	commentSection = indent(commentSection)
	commentSection = moveZeroWidthSpaceUpOneLine(commentSection)

	return commentSection
}

func moveZeroWidthSpaceUpOneLine(commentSection string) string {
	indentBlock := getIndentBlock()

	return strings.ReplaceAll(commentSection, newLine+indentBlock+unicode.ZeroWidthSpace,
		unicode.ZeroWidthSpace+newLine+indentBlock)
}

func indent(commentSection string) string {
	indentBlock := getIndentBlock()
	lines := strings.Split(commentSection, "\n")
	indentedCommentSection := ""

	for _, line := range lines {
		indentedCommentSection += indentBlock + line + "\n"
	}

	return indentedCommentSection
}

func getIndentBlock() string {
	return strings.Repeat(" ", margins.CommentSectionLeftMargin)
}
