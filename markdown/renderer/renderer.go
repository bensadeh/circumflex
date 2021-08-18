package renderer

import "clx/markdown"

func ToString(blocks []*markdown.Block) string {
	output := ""

	for _, block := range blocks {
		switch {
		case block.Kind == markdown.Text:
			output += renderText(block.Text) + "\n\n"

		default:
			output += renderText(block.Text) + "\n\n"
		}
	}

	return output
}

func renderText(text string) string {
	return text
}
