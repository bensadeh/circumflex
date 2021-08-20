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

	return output
}

func renderCode(text string) string {
	text = strings.TrimSuffix(text, "\n")
	text = strings.TrimPrefix(text, "\n")

	return text
}

func renderQuote(text string, lineWidth int, altIndentBlock bool) string {
	indentSymbol := " " + indent.GetIndentSymbol(false, altIndentBlock)
	padding := termtext.WrapPad(Faint(indentSymbol).String())
	text, _ = termtext.Wrap(text, 70, padding)

	// text = strings.TrimSuffix(text, "\n")
	// text = strings.TrimPrefix(text, "\n")

	return text
}
