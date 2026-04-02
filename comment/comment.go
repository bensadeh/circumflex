package comment

import (
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/syntax"

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

func Render(commentHTML string, commentWidth int, availableScreenWidth int, enableNerdFonts bool) string {
	if commentHTML == "[deleted]" {
		return style.Faint(commentHTML)
	}

	sections := parseSections(commentHTML)

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

func formatQuote(content string, commentWidth int) string {
	content = strings.ReplaceAll(content, "<i>", "")
	content = strings.ReplaceAll(content, "</i>", "")
	content = strings.ReplaceAll(content, "</a>", ansi.Reset+ansi.Faint+ansi.Italic)
	content = syntax.ReplaceSymbols(content)
	content = syntax.ConvertSmileys(content)

	content = strings.Replace(content, ">>", "", 1)
	content = strings.Replace(content, ">", "", 1)
	content = strings.TrimLeft(content, " ")
	content = syntax.TrimURLs(content, true)
	content = syntax.RemoveUnwantedNewLines(content)
	content = syntax.RemoveUnwantedWhitespace(content)

	content = ansi.Italic + ansi.Faint + content + ansi.Reset

	quoteIndent := " " + style.IndentSymbol
	padding := text.WrapPad(ansi.Faint + quoteIndent)
	wrapped, _ := text.Wrap(content, commentWidth, padding)

	return wrapped
}

func formatCodeBlock(content string, availableWidth int) string {
	content = syntax.ReplaceHTML(content)

	wrapped, _ := text.Wrap(content, availableWidth)
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

func formatParagraph(content string, commentWidth int, enableNerdFonts bool) string {
	content = syntax.ReplaceSymbols(content)
	content = syntax.ConvertSmileys(content)
	content = syntax.ReplaceHTML(content)
	content = strings.TrimLeft(content, " ")
	content = highlightCommentSyntax(content, enableNerdFonts)
	content = syntax.TrimURLs(content, true)
	content = syntax.RemoveUnwantedNewLines(content)
	content = syntax.RemoveUnwantedWhitespace(content)

	wrapped, _ := text.Wrap(content, commentWidth)

	return wrapped
}

func isQuote(s string) bool {
	quoteMark := ">"

	return strings.HasPrefix(s, quoteMark) ||
		strings.HasPrefix(s, " "+quoteMark) ||
		strings.HasPrefix(s, "<i>"+quoteMark) ||
		strings.HasPrefix(s, "<i> "+quoteMark)
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
