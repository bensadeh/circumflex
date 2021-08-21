package parser

import (
	"clx/markdown"
	"errors"
	"strings"
)

func Parse(text string) []*markdown.Block {
	var blocks []*markdown.Block

	lines := strings.Split(text+"\n", "\n")
	temp := new(tempBuffer)

	isInsideQuote := false
	isInsideCode := false
	isInsideText := false
	isInsideList := false
	isInsideTable := false

	for _, line := range lines {
		if isInsideCode {
			if strings.HasPrefix(line, "```") {
				isInsideCode = false

				appendedBlocks, err := appendNonEmptyBuffer(temp, blocks)
				if err == nil {
					blocks = appendedBlocks
				}

				temp.reset()

				continue
			}

			temp.append("\n" + line)

			continue
		}

		// isAtLastLine := i == len(lines)-1

		if line == "" {
			appendedBlocks, err := appendNonEmptyBuffer(temp, blocks)
			if err == nil {
				blocks = appendedBlocks
			}

			temp.reset()

			isInsideQuote = false
			isInsideText = false
			isInsideList = false
			isInsideTable = false

			continue
		}

		if isInsideTable {
			temp.append("\n" + line)

			continue
		}

		if isInsideText {
			temp.append(" " + line)

			continue
		}

		if isInsideQuote || isInsideList {
			line = strings.TrimPrefix(line, ">")
			line = strings.TrimPrefix(line, " ")

			temp.append("\n" + line)

			continue
		}

		switch {
		case strings.HasPrefix(line, `![`):
			temp.kind = markdown.Image
			temp.text = line

		case strings.HasPrefix(line, "> "):
			temp.kind = markdown.Quote
			temp.text = strings.TrimPrefix(line, "> ")

			isInsideQuote = true

		case strings.HasPrefix(line, "```"):
			temp.kind = markdown.Code
			temp.text = ""

			isInsideCode = true

		case isListItem(line):
			temp.kind = markdown.List
			temp.text = line

			isInsideList = true

		case strings.HasPrefix(line, "|"):
			temp.kind = markdown.Table
			temp.text = line

			isInsideTable = true

		case strings.HasPrefix(line, "# "):
			temp.kind = markdown.H1
			temp.text = line

			isInsideText = true

		case strings.HasPrefix(line, "## "):
			temp.kind = markdown.H2
			temp.text = line

			isInsideText = true

		case strings.HasPrefix(line, "### "):
			temp.kind = markdown.H3
			temp.text = line

			isInsideText = true

		case strings.HasPrefix(line, "#### "):
			temp.kind = markdown.H4
			temp.text = line

			isInsideText = true

		case strings.HasPrefix(line, "##### "):
			temp.kind = markdown.H5
			temp.text = line

			isInsideText = true

		case strings.HasPrefix(line, "###### "):
			temp.kind = markdown.H6
			temp.text = line

			isInsideText = true

		default:
			temp.kind = markdown.Text
			temp.text = line

			isInsideText = true
		}
	}

	return blocks
}

func isListItem(text string) bool {
	if strings.HasPrefix(text, "- ") || strings.HasPrefix(text, "1. ") {
		return true
	}

	return false
}

func appendNonEmptyBuffer(temp *tempBuffer, blocks []*markdown.Block) ([]*markdown.Block, error) {
	if temp.kind == markdown.Text && temp.text == "" {
		return nil, errors.New("buffer is empty")
	}

	b := markdown.Block{
		Kind: temp.kind,
		Text: temp.text,
	}

	return append(blocks, &b), nil
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
