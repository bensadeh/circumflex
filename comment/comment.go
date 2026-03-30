package comment

import (
	"clx/ansi"
	"clx/settings"
	"clx/style"
	"clx/syntax"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

type comment struct {
	sections []*section
}

type section struct {
	isCodeBlock bool
	isQuote     bool
	content     string
}

func Print(c string, config *settings.Config, commentWidth int, availableScreenWidth int) string {
	if c == "[deleted]" {
		return style.Faint(c)
	}

	c = strings.TrimPrefix(c, "<p>")
	c = strings.ReplaceAll(c, "\n</code></pre>\n", "<p>")
	paragraphs := strings.Split(c, "<p>")

	comment := new(comment)
	comment.sections = make([]*section, len(paragraphs))

	for i, paragraph := range paragraphs {
		s := new(section)
		s.content = syntax.ReplaceCharacters(paragraph)

		if strings.Contains(s.content, "<pre><code>") {
			s.isCodeBlock = true
		}

		if isQuote(s.content) {
			s.isQuote = true
		}

		comment.sections[i] = s
	}

	var output strings.Builder

	for i, s := range comment.sections {
		paragraph := s.content

		switch {
		case s.isQuote:
			paragraph = strings.ReplaceAll(paragraph, "<i>", "")
			paragraph = strings.ReplaceAll(paragraph, "</i>", "")
			paragraph = strings.ReplaceAll(paragraph, "</a>", ansi.Reset+ansi.Faint+ansi.Italic)
			paragraph = syntax.ReplaceSymbols(paragraph)
			paragraph = convertToEmojis(paragraph, config.DisableEmojis)

			paragraph = strings.Replace(paragraph, ">>", "", 1)
			paragraph = strings.Replace(paragraph, ">", "", 1)
			paragraph = strings.TrimLeft(paragraph, " ")
			paragraph = syntax.TrimURLs(paragraph, false)
			paragraph = syntax.RemoveUnwantedNewLines(paragraph)
			paragraph = syntax.RemoveUnwantedWhitespace(paragraph)

			paragraph = ansi.Italic + ansi.Faint + paragraph + ansi.Reset

			quoteIndent := " " + config.IndentationSymbol
			padding := text.WrapPad(ansi.Faint + quoteIndent)
			wrappedAndPaddedComment, _ := text.Wrap(paragraph, commentWidth, padding)
			paragraph = wrappedAndPaddedComment

		case s.isCodeBlock:
			paragraph = syntax.ReplaceHTML(paragraph)
			wrappedComment, _ := text.Wrap(paragraph, availableScreenWidth)

			codeLines := strings.Split(wrappedComment, "\n")

			var formattedCodeLines strings.Builder

			for j, codeLine := range codeLines {
				isOnLastLine := j == len(codeLines)-1

				if isOnLastLine {
					formattedCodeLines.WriteString(ansi.Faint + codeLine + ansi.Reset)

					break
				}

				formattedCodeLines.WriteString(ansi.Faint + codeLine + ansi.Reset + "\n")
			}

			paragraph = formattedCodeLines.String()

		default:
			paragraph = syntax.ReplaceSymbols(paragraph)
			paragraph = convertToEmojis(paragraph, config.DisableEmojis)

			paragraph = syntax.ReplaceHTML(paragraph)
			paragraph = strings.TrimLeft(paragraph, " ")
			paragraph = highlightCommentSyntax(paragraph, config.DisableCommentHighlighting, config.EnableNerdFonts)

			paragraph = syntax.TrimURLs(paragraph, config.DisableCommentHighlighting)
			paragraph = syntax.RemoveUnwantedNewLines(paragraph)
			paragraph = syntax.RemoveUnwantedWhitespace(paragraph)

			wrappedAndPaddedComment, _ := text.Wrap(paragraph, commentWidth)
			paragraph = wrappedAndPaddedComment
		}

		separator := getParagraphSeparator(i, len(comment.sections))
		output.WriteString(paragraph + separator)
	}

	return output.String()
}

func convertToEmojis(paragraph string, disableEmojis bool) string {
	if disableEmojis {
		return paragraph
	}

	paragraph = syntax.ConvertSmileys(paragraph)

	return paragraph
}

func isQuote(text string) bool {
	quoteMark := ">"

	return strings.HasPrefix(text, quoteMark) ||
		strings.HasPrefix(text, " "+quoteMark) ||
		strings.HasPrefix(text, "<i>"+quoteMark) ||
		strings.HasPrefix(text, "<i> "+quoteMark)
}

func getParagraphSeparator(index int, sliceLength int) string {
	isAtLastParagraph := index == sliceLength-1

	if isAtLastParagraph {
		return ""
	}

	return "\n\n"
}

func highlightCommentSyntax(input string, disableCommentHighlighting bool, enableNerdFonts bool) string {
	if disableCommentHighlighting {
		return input
	}

	input = syntax.HighlightBackticks(input)
	input = syntax.HighlightMentions(input)
	input = syntax.HighlightVariables(input)
	input = syntax.HighlightAbbreviations(input)
	input = syntax.HighlightReferences(input)
	input = syntax.HighlightYCStartupsInHeadlines(input, syntax.Unselected, enableNerdFonts)

	return input
}
