package comment

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/syntax"
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

func Render(commentHTML string, commentWidth, screenWidth int, enableNerdFonts bool, fg color.Color) string {
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
			output.WriteString(formatCodeBlock(s.content, commentWidth, screenWidth))
		case sectionParagraph:
			para := formatParagraph(s.content, commentWidth, enableNerdFonts)

			if fg != nil {
				para = style.PaintForeground(para, fg)
			}

			output.WriteString(para)
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
	padStr := ansi.Faint + quoteIndent
	padWidth := lipgloss.Width(padStr)
	wrapped := lipgloss.Wrap(content, commentWidth-padWidth, "")

	return style.PrefixLines(wrapped, padStr)
}

// The box spans at least commentWidth and grows with long code lines up to
// screenWidth (the space left of the scrollbar).
func formatCodeBlock(content string, commentWidth, screenWidth int) string {
	content = syntax.ReplaceHTML(content)
	content = dedent(strings.Trim(content, "\n"))

	wrapped := lipgloss.Wrap(content, screenWidth-style.RoundedBoxChrome, "")
	lines := strings.Split(wrapped, "\n")

	for i, line := range lines {
		lines[i] = ansi.Faint + line + ansi.Reset
	}

	return style.RoundedBox(strings.Join(lines, "\n"), commentWidth)
}

// dedent drops the whitespace indent shared by every line. HN marks code by
// leading spaces, so the block always arrives indented; the box already sets
// it apart, and keeping the indent would pad the box's inside.
func dedent(text string) string {
	lines := strings.Split(text, "\n")

	indent := -1

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if trimmed == "" {
			continue
		}

		if lineIndent := len(line) - len(trimmed); indent < 0 || lineIndent < indent {
			indent = lineIndent
		}
	}

	if indent <= 0 {
		return text
	}

	for i, line := range lines {
		if len(line) >= indent {
			lines[i] = line[indent:]
		}
	}

	return strings.Join(lines, "\n")
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

	return lipgloss.Wrap(content, commentWidth, "")
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
