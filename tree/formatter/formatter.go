package formatter

import (
	"clx/constants"
	"clx/reader"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

const (
	newLine = "\n"
)

func Process(commentSection string, screenWidth int) string {
	commentSection = indentCommentSection(commentSection, screenWidth)
	commentSection = moveZeroWidthSpaceUpOneLine(commentSection)
	commentSection = reader.DeIndentInfoSection(commentSection)

	return commentSection
}

func moveZeroWidthSpaceUpOneLine(commentSection string) string {
	indentBlock := getIndentBlock()

	return strings.ReplaceAll(commentSection, newLine+indentBlock+constants.InvisibleCharacterForTopLevelComments,
		constants.InvisibleCharacterForTopLevelComments+newLine+indentBlock)
}

func indentCommentSection(commentSection string, screenWidth int) string {
	indentBlock := getIndentBlock()

	indentedCommentSection, _ := text.WrapWithPad(commentSection, screenWidth, indentBlock)

	return indentedCommentSection
}

func getIndentBlock() string {
	return strings.Repeat(" ", constants.CommentSectionLeftMargin)
}
