package comment

import (
	"clx/ansi"
	"clx/settings"
	"clx/style"
	"clx/syntax"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

type sectionKind int

const (
	sectionParagraph sectionKind = iota
	sectionCode
	sectionQuote
)

type section struct {
	kind    sectionKind
	content string
}

func Print(c string, config *settings.Config, commentWidth int, availableScreenWidth int) string {
	if c == "[deleted]" {
		return style.Faint(c)
	}

	sections := parseSections(c)

	var output strings.Builder

	for i, s := range sections {
		switch s.kind {
		case sectionQuote:
			output.WriteString(formatQuote(s.content, config, commentWidth))
		case sectionCode:
			output.WriteString(formatCodeBlock(s.content, availableScreenWidth))
		case sectionParagraph:
			output.WriteString(formatParagraph(s.content, config, commentWidth))
		}

		if i < len(sections)-1 {
			output.WriteString("\n\n")
		}
	}

	return output.String()
}

func parseSections(html string) []section {
	html = strings.TrimPrefix(html, "<p>")
	html = strings.ReplaceAll(html, "\n</code></pre>\n", "<p>")
	paragraphs := strings.Split(html, "<p>")

	sections := make([]section, 0, len(paragraphs))

	for _, p := range paragraphs {
		content := syntax.ReplaceCharacters(p)

		kind := sectionParagraph

		switch {
		case strings.Contains(content, "<pre><code>"):
			kind = sectionCode
		case isQuote(content):
			kind = sectionQuote
		}

		sections = append(sections, section{kind: kind, content: content})
	}

	return sections
}

func formatQuote(paragraph string, config *settings.Config, commentWidth int) string {
	paragraph = strings.ReplaceAll(paragraph, "<i>", "")
	paragraph = strings.ReplaceAll(paragraph, "</i>", "")
	paragraph = strings.ReplaceAll(paragraph, "</a>", ansi.Reset+ansi.Faint+ansi.Italic)
	paragraph = syntax.ReplaceSymbols(paragraph)
	paragraph = convertToEmojis(paragraph, config.DisableEmojis)

	paragraph = strings.Replace(paragraph, ">>", "", 1)
	paragraph = strings.Replace(paragraph, ">", "", 1)
	paragraph = strings.TrimLeft(paragraph, " ")
	paragraph = syntax.TrimURLs(paragraph, true)
	paragraph = syntax.RemoveUnwantedNewLines(paragraph)
	paragraph = syntax.RemoveUnwantedWhitespace(paragraph)

	paragraph = ansi.Italic + ansi.Faint + paragraph + ansi.Reset

	quoteIndent := " " + config.IndentationSymbol
	padding := text.WrapPad(ansi.Faint + quoteIndent)
	wrapped, _ := text.Wrap(paragraph, commentWidth, padding)

	return wrapped
}

func formatCodeBlock(paragraph string, availableWidth int) string {
	paragraph = syntax.ReplaceHTML(paragraph)

	wrapped, _ := text.Wrap(paragraph, availableWidth)
	lines := strings.Split(wrapped, "\n")

	var sb strings.Builder

	for i, line := range lines {
		sb.WriteString(ansi.Faint + line + ansi.Reset)

		if i < len(lines)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func formatParagraph(paragraph string, config *settings.Config, commentWidth int) string {
	paragraph = syntax.ReplaceSymbols(paragraph)
	paragraph = convertToEmojis(paragraph, config.DisableEmojis)
	paragraph = syntax.ReplaceHTML(paragraph)
	paragraph = strings.TrimLeft(paragraph, " ")
	paragraph = highlightCommentSyntax(paragraph, config.EnableNerdFonts)
	paragraph = syntax.TrimURLs(paragraph, true)
	paragraph = syntax.RemoveUnwantedNewLines(paragraph)
	paragraph = syntax.RemoveUnwantedWhitespace(paragraph)

	wrapped, _ := text.Wrap(paragraph, commentWidth)

	return wrapped
}

func convertToEmojis(paragraph string, disableEmojis bool) string {
	if disableEmojis {
		return paragraph
	}

	return syntax.ConvertSmileys(paragraph)
}

func isQuote(text string) bool {
	quoteMark := ">"

	return strings.HasPrefix(text, quoteMark) ||
		strings.HasPrefix(text, " "+quoteMark) ||
		strings.HasPrefix(text, "<i>"+quoteMark) ||
		strings.HasPrefix(text, "<i> "+quoteMark)
}

func highlightCommentSyntax(input string, enableNerdFonts bool) string {
	input = syntax.HighlightBackticks(input)
	input = syntax.HighlightMentions(input)
	input = syntax.HighlightVariables(input)
	input = syntax.HighlightAbbreviations(input)
	input = syntax.HighlightReferences(input)
	input = syntax.HighlightYCStartupsInHeadlines(input, syntax.Unselected, enableNerdFonts)

	return input
}
