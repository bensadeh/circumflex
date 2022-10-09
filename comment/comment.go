package comment

import (
	"strings"

	"clx/settings"
	"clx/syntax"

	"github.com/logrusorgru/aurora/v3"

	text "github.com/MichaelMure/go-term-text"
)

const (
	reset  = "\033[0m"
	dimmed = "\033[2m"
	italic = "\033[3m"
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
		return aurora.Faint(c).String()
	}

	c = strings.Replace(c, "<p>", "", 1)
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

	output := ""

	for i, s := range comment.sections {
		paragraph := s.content

		switch {
		case s.isQuote:
			paragraph = strings.ReplaceAll(paragraph, "<i>", "")
			paragraph = strings.ReplaceAll(paragraph, "</i>", "")
			paragraph = strings.ReplaceAll(paragraph, "</a>", reset+dimmed+italic)
			paragraph = syntax.ReplaceSymbols(paragraph)
			paragraph = convertToEmojis(paragraph, config.DisableEmojis)

			paragraph = strings.Replace(paragraph, ">>", "", 1)
			paragraph = strings.Replace(paragraph, ">", "", 1)
			paragraph = strings.TrimLeft(paragraph, " ")
			paragraph = syntax.TrimURLs(paragraph, false)
			paragraph = syntax.RemoveUnwantedNewLines(paragraph)
			paragraph = syntax.RemoveUnwantedWhitespace(paragraph)

			paragraph = italic + dimmed + paragraph + reset

			quoteIndent := " " + config.IndentationSymbol
			padding := text.WrapPad(dimmed + quoteIndent)
			wrappedAndPaddedComment, _ := text.Wrap(paragraph, commentWidth, padding)
			paragraph = wrappedAndPaddedComment

		case s.isCodeBlock:
			paragraph = syntax.ReplaceHTML(paragraph)
			wrappedComment, _ := text.Wrap(paragraph, availableScreenWidth)

			codeLines := strings.Split(wrappedComment, "\n")
			formattedCodeLines := ""

			for j, codeLine := range codeLines {
				isOnLastLine := j == len(codeLines)-1

				if isOnLastLine {
					formattedCodeLines += dimmed + codeLine + reset

					break
				}

				formattedCodeLines += dimmed + codeLine + reset + "\n"
			}

			paragraph = formattedCodeLines

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
		output += paragraph + separator
	}

	return output
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
