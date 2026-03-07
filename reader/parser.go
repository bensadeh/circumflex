package reader

import (
	"errors"
	"regexp"
	"strings"
)

const (
	enDash     = "–"
	emDash     = "—"
	normalDash = "-"
)

func convertToMarkdownBlocks(text string) []*block {
	var blocks []*block

	text = strings.ReplaceAll(text, enDash, normalDash)
	text = strings.ReplaceAll(text, emDash, normalDash)

	text = strings.ReplaceAll(text, boldStart, "")
	text = strings.ReplaceAll(text, boldStop, "")

	lines := strings.Split(text+"\n", "\n")
	temp := new(tempBuffer)

	isInsideQuote := false
	isInsideCode := false
	isInsideText := false
	isInsideList := false
	isInsideTable := false

	for _, line := range lines {
		lineWithoutFormatting := strings.TrimLeft(line, " ")
		lineWithoutFormatting = strings.ReplaceAll(lineWithoutFormatting, boldStart, "")
		lineWithoutFormatting = strings.ReplaceAll(lineWithoutFormatting, italicStart, "")

		if isInsideCode {
			if strings.HasPrefix(lineWithoutFormatting, "```") {
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

		if isInsideList {
			temp.append("\n" + line)

			continue
		}

		if isInsideQuote {
			line = strings.TrimPrefix(line, ">")
			line = strings.TrimPrefix(line, " ")

			temp.append("\n" + line)

			continue
		}

		switch {
		case strings.HasPrefix(lineWithoutFormatting, `![`):
			temp.kind = blockImage
			temp.text = line

		case strings.HasPrefix(lineWithoutFormatting, "> "):
			temp.kind = blockQuote
			temp.text = strings.TrimPrefix(line, "> ")

			isInsideQuote = true

		case strings.HasPrefix(lineWithoutFormatting, "```"):
			temp.kind = blockCode
			temp.text = ""

			isInsideCode = true

		case isListItem(lineWithoutFormatting):
			if isSameTypeAsPreviousItem(blockList, blocks) {
				lastItem := len(blocks) - 1

				temp.kind = blockList
				temp.text = blocks[lastItem].Text + "\n" + line

				blocks = removeIndex(blocks, lastItem)
				isInsideList = true

				continue
			}

			temp.kind = blockList
			temp.text = line

			isInsideList = true

		case strings.HasPrefix(lineWithoutFormatting, "|"):
			if isSameTypeAsPreviousItem(blockTable, blocks) {
				lastItem := len(blocks) - 1

				temp.kind = blockTable
				temp.text = blocks[lastItem].Text + "\n" + line

				blocks = removeIndex(blocks, lastItem)
				isInsideTable = true

				continue
			}

			temp.kind = blockTable
			temp.text = line

			isInsideTable = true

		case strings.HasPrefix(lineWithoutFormatting, "* * *"):
			temp.kind = blockDivider
			temp.text = line

		case strings.HasPrefix(lineWithoutFormatting, "# "):
			temp.kind = blockH1
			temp.text = lineWithoutFormatting

			isInsideText = true

		case strings.HasPrefix(lineWithoutFormatting, "## "):
			temp.kind = blockH2
			temp.text = lineWithoutFormatting

			isInsideText = true

		case strings.HasPrefix(lineWithoutFormatting, "### "):
			temp.kind = blockH3
			temp.text = lineWithoutFormatting

			isInsideText = true

		case strings.HasPrefix(lineWithoutFormatting, "#### "):
			temp.kind = blockH4
			temp.text = lineWithoutFormatting

			isInsideText = true

		case strings.HasPrefix(lineWithoutFormatting, "##### "):
			temp.kind = blockH5
			temp.text = lineWithoutFormatting

			isInsideText = true

		case strings.HasPrefix(lineWithoutFormatting, "###### "):
			temp.kind = blockH6
			temp.text = lineWithoutFormatting

			isInsideText = true

		default:
			temp.kind = blockText
			temp.text = line

			isInsideText = true
		}
	}

	return blocks
}

func removeIndex(s []*block, index int) []*block {
	return append(s[:index], s[index+1:]...)
}

func isListItem(text string) bool {
	if text == "" {
		return false
	}

	exp := regexp.MustCompile(`^\s*(-|\d+\. )`)
	listToken := exp.FindString(text)

	return listToken != ""
}

func isSameTypeAsPreviousItem(itemType int, blocks []*block) bool {
	if len(blocks) == 0 {
		return false
	}

	previousItem := len(blocks) - 1

	return blocks[previousItem].Kind == itemType
}

func appendNonEmptyBuffer(temp *tempBuffer, blocks []*block) ([]*block, error) {
	if temp.kind == blockText && temp.text == "" {
		return nil, errors.New("buffer is empty")
	}

	b := block{
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
