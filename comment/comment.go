package comment

import (
	"clx/ansi"
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

func Print(c string, commentWidth int, availableScreenWidth int, enableNerdFonts bool) string {
	if c == "[deleted]" {
		return style.Faint(c)
	}

	sections := parseSections(c)

	var output strings.Builder

	for i, s := range sections {
		switch s.kind {
		case sectionQuote:
			output.WriteString(formatQuote(s.content, commentWidth))
		case sectionCode:
			output.WriteString(formatCodeBlock(s.content, availableScreenWidth))
		case sectionParagraph:
			output.WriteString(formatParagraph(s.content, commentWidth, enableNerdFonts))
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

func formatQuote(paragraph string, commentWidth int) string {
	paragraph = strings.ReplaceAll(paragraph, "<i>", "")
	paragraph = strings.ReplaceAll(paragraph, "</i>", "")
	paragraph = strings.ReplaceAll(paragraph, "</a>", ansi.Reset+ansi.Faint+ansi.Italic)
	paragraph = syntax.ReplaceSymbols(paragraph)
	paragraph = syntax.ConvertSmileys(paragraph)

	paragraph = strings.Replace(paragraph, ">>", "", 1)
	paragraph = strings.Replace(paragraph, ">", "", 1)
	paragraph = strings.TrimLeft(paragraph, " ")
	paragraph = syntax.TrimURLs(paragraph, true)
	paragraph = syntax.RemoveUnwantedNewLines(paragraph)
	paragraph = syntax.RemoveUnwantedWhitespace(paragraph)

	paragraph = ansi.Italic + ansi.Faint + paragraph + ansi.Reset

	quoteIndent := " " + style.IndentSymbol
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

func formatParagraph(paragraph string, commentWidth int, enableNerdFonts bool) string {
	paragraph = syntax.ReplaceSymbols(paragraph)
	paragraph = syntax.ConvertSmileys(paragraph)
	paragraph = syntax.ReplaceHTML(paragraph)
	paragraph = strings.TrimLeft(paragraph, " ")
	paragraph = highlightCommentSyntax(paragraph, enableNerdFonts)
	paragraph = syntax.TrimURLs(paragraph, true)
	paragraph = syntax.RemoveUnwantedNewLines(paragraph)
	paragraph = syntax.RemoveUnwantedWhitespace(paragraph)

	wrapped, _ := text.Wrap(paragraph, commentWidth)

	return wrapped
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
