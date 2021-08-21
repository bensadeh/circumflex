package renderer

import (
	"clx/indent"
	"clx/markdown"
	"clx/syntax"
	"regexp"
	"strings"

	termtext "github.com/MichaelMure/go-term-text"
	. "github.com/logrusorgru/aurora/v3"
)

func ToString(blocks []*markdown.Block, lineWidth int, altIndentBlock bool) string {
	output := ""

	for _, block := range blocks {
		switch block.Kind {
		case markdown.Text:
			output += renderText(block.Text) + "\n\n"

		case markdown.Image:
			output += renderImage(block.Text) + "\n\n"

		case markdown.Code:
			output += renderCode(block.Text) + "\n\n"

		case markdown.Quote:
			output += renderQuote(block.Text, 80, altIndentBlock) + "\n\n"

		case markdown.H1:
			output += h1(block.Text) + "\n\n"

		case markdown.H2:
			output += h2(block.Text) + "\n\n"

		case markdown.H3:
			output += h3(block.Text) + "\n\n"

		case markdown.H4:
			output += h4(block.Text) + "\n\n"

		case markdown.H5:
			output += h1(block.Text) + "\n\n"

		case markdown.H6:
			output += h1(block.Text) + "\n\n"

		default:
			output += renderText(block.Text) + "\n\n"
		}
	}

	return output
}

func renderText(text string) string {
	text = it(text)
	text = bld(text)

	text = syntax.HighlightBackticks(text)

	padding := termtext.WrapPad("")
	text, _ = termtext.Wrap(text, 80, padding)

	return text
}

func renderImage(text string) string {
	exp := regexp.MustCompile(`!\[(.*?)\]\(.*?\)`)
	imageLabel := Magenta("Image: %%%").Faint().String()

	text = exp.ReplaceAllString(text, imageLabel+Italic(`$1.`).Faint().String()+"### ")

	lines := strings.Split(text, imageLabel)
	output := ""

	for i, line := range lines {
		if i == 0 || line == "" {
			continue
		}

		output += imageLabel + line + "\n\n"
	}

	// Remove ': ' if there is no image caption
	output = strings.ReplaceAll(output, ": %%%\u001B[0m\u001B[3m.\u001B[0m###", "\u001B[0m ")

	output = strings.ReplaceAll(output, "###", "")
	output = strings.ReplaceAll(output, "%%%", "")
	output = strings.TrimSuffix(output, "\n\n")

	output = it(output)
	output = bld(output)

	return output
}

func renderCode(text string) string {
	text = strings.TrimSuffix(text, "\n")
	text = strings.TrimPrefix(text, "\n")

	return text
}

func renderQuote(text string, lineWidth int, altIndentBlock bool) string {
	text = Italic(text).String()

	indentSymbol := " " + indent.GetIndentSymbol(false, altIndentBlock)
	padding := termtext.WrapPad(Faint(indentSymbol).String())
	text = itReversed(text)
	text = bldInQuote(text)
	text, _ = termtext.Wrap(text, 70, padding)

	// text = strings.TrimSuffix(text, "\n")
	// text = strings.TrimPrefix(text, "\n")

	return text
}

func it(text string) string {
	italic := "\u001B[3m"
	noItalic := "\u001B[23m"

	text = strings.ReplaceAll(text, markdown.ItalicStart, italic)
	text = strings.ReplaceAll(text, markdown.ItalicStop, noItalic)

	return text
}

func itReversed(text string) string {
	italic := "\u001B[3m"
	noItalic := "\u001B[23m"

	text = strings.ReplaceAll(text, markdown.ItalicStart, noItalic)
	text = strings.ReplaceAll(text, markdown.ItalicStop, italic)

	return text
}

func bld(text string) string {
	bold := "\033[31m"
	noBold := "\033[0m"

	text = strings.ReplaceAll(text, markdown.BoldStart, bold)
	text = strings.ReplaceAll(text, markdown.BoldStop, noBold)

	return text
}

func bldInQuote(text string) string {
	// bold := "\033[31m"
	// noBold := "\033[0m"

	text = strings.ReplaceAll(text, markdown.BoldStart, "")
	text = strings.ReplaceAll(text, markdown.BoldStop, "")

	return text
}

func h1(text string) string {
	text = strings.TrimPrefix(text, "# ")

	return Bold(text).String()
}

func h2(text string) string {
	text = strings.TrimPrefix(text, "## ")

	return Bold(text).Blue().String()
}

func h3(text string) string {
	text = strings.TrimPrefix(text, "### ")

	return Bold(text).Yellow().String()
}

func h4(text string) string {
	text = strings.TrimPrefix(text, "#### ")

	return Bold(text).Green().String()
}
