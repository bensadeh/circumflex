package reader

import (
	"clx/constants"
	"os"
	"strings"

	t "github.com/MichaelMure/go-term-text"
	"golang.org/x/term"
)

const (
	newLine            = "\n"
	defaultScreenWidth = 80
)

func processArticle(text string, url string) string {
	text = filterSite(text, url)
	text = moveZeroWidthSpaceUpOneLine(text)
	text = indent(text)
	text = DeIndentInfoSection(text)

	return text
}

func moveZeroWidthSpaceUpOneLine(text string) string {
	return strings.ReplaceAll(text, newLine+constants.InvisibleCharacterForTopLevelComments,
		constants.InvisibleCharacterForTopLevelComments+newLine)
}

func indent(commentSection string) string {
	indentBlock := strings.Repeat(" ", constants.ReaderViewLeftMargin)

	screenWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		screenWidth = defaultScreenWidth
	}

	indentedCommentSection, _ := t.WrapWithPad(commentSection, screenWidth, indentBlock)

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
