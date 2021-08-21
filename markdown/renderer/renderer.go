package renderer

import (
	"clx/indent"
	"clx/markdown"
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

		default:
			output += renderText(block.Text) + "\n\n"
		}
	}

	return output
}

func renderText(text string) string {
	text = it(text)
	text = bld(text)

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

	expStart := regexp.MustCompile(`\*\*(\w{1})`)
	text = expStart.ReplaceAllString(text, bold+`$1`)

	expEnd := regexp.MustCompile(`([a-zA-Z0-9.)%])\*\*`)
	text = expEnd.ReplaceAllString(text, `$1`+noBold)

	return text
}
