package parser

import (
	"clx/markdown"
	"strings"
)

func Parse(text string) []*markdown.Block {
	lines := strings.Split(text, "\n")

	var blocks []*markdown.Block
	temp := new(tempBuffer)

	isInsideQuote := false
	isInsideText := false

	for _, line := range lines {
		if line == "" {
			b := markdown.Block{
				Kind: temp.kind,
				Text: temp.text,
			}

			blocks = append(blocks, &b)

			temp.reset()

			isInsideQuote = false
			isInsideText = false

			continue
		}

		if isInsideText {
			temp.append(" " + line)

			continue
		}

		if isInsideQuote {
			temp.append(line)

			continue
		}

		switch {
		case strings.HasPrefix(line, "!["):
			temp.kind = markdown.Image
			temp.text = line

		case strings.HasPrefix(line, "> "):
			temp.kind = markdown.Quote
			temp.text = line

			isInsideQuote = true

		default:
			temp.kind = markdown.Text
			temp.text = line

			isInsideText = true
		}
	}

	return blocks
}

type tempBuffer struct {
	kind int
	text string
}

func (b *tempBuffer) reset() {
	b.kind = 0
	b.text = ""
}

func (b *tempBuffer) append(text string) {
	b.text += text
}
