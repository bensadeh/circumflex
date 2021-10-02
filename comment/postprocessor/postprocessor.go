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
	commentSection = indent(commentSection, screenWidth)
	commentSection = moveZeroWidthSpaceUpOneLine(commentSection)

	return commentSection
}

func moveZeroWidthSpaceUpOneLine(commentSection string) string {
	indentBlock := getIndentBlock()

	return strings.ReplaceAll(commentSection, newLine+indentBlock+unicode.ZeroWidthSpace,
		unicode.ZeroWidthSpace+newLine+indentBlock)
}

func indent(commentSection string, screenWidth int) string {
	indentBlock := getIndentBlock()

	indentedCommentSection, _ := text.WrapWithPad(commentSection, screenWidth, indentBlock)

	return indentedCommentSection
}

func getIndentBlock() string {
	return strings.Repeat(" ", margins.CommentSectionLeftMargin)
}
