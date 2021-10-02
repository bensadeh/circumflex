package postprocessor

import (
	"clx/constants/margins"
	"clx/constants/unicode"
	"clx/screen"
	"strings"

	t "github.com/MichaelMure/go-term-text"
)

const (
	newLine = "\n"
)

func Process(text string, url string) string {
	text = filterSite(text, url)
	text = moveZeroWidthSpaceUpOneLine(text)
	text = indent(text)

	return text
}

func moveZeroWidthSpaceUpOneLine(text string) string {
	return strings.ReplaceAll(text, newLine+unicode.ZeroWidthSpace,
		unicode.ZeroWidthSpace+newLine)
}

func indent(commentSection string) string {
	indentBlock := strings.Repeat(" ", margins.ReaderViewLeftMargin)
	screenWidth := screen.GetTerminalWidth() - margins.ReaderViewLeftMargin

	indentedCommentSection, _ := t.WrapWithPad(commentSection, screenWidth, indentBlock)

	return indentedCommentSection
}
