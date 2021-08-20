package renderer

import (
	"clx/markdown"
	"regexp"
	"strings"

	. "github.com/logrusorgru/aurora/v3"
)

func ToString(blocks []*markdown.Block) string {
	output := ""

	for _, block := range blocks {
		switch {
		case block.Kind == markdown.Text:
			output += renderText(block.Text) + "\n\n"

		case block.Kind == markdown.Image:
			output += renderImage(block.Text) + "\n\n"

		case block.Kind == markdown.Code:
			output += renderCode(block.Text) + "\n\n"

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
