package article

import "strings"

type blockKind int

const (
	blockParagraph blockKind = iota
	blockHeading
	blockList
	blockQuote
	blockCode
	blockTable
	blockImage
	blockDivider
)

type block struct {
	kind  blockKind
	level int        // blockHeading: 1-6
	spans []span     // blockParagraph, blockQuote, blockImage (caption)
	items []listItem // blockList
	rows  [][]string // blockTable, first row is the header
	text  string     // blockHeading, blockCode
}

type inlineFormat int

const (
	formatPlain inlineFormat = iota
	formatItalic
	formatCode
)

type span struct {
	text   string
	format inlineFormat
}

type listItem struct {
	depth  int
	number int // 1-based position for ordered items, 0 for bullets
	spans  []span
}

func spanText(spans []span) string {
	var sb strings.Builder
	for _, s := range spans {
		sb.WriteString(s.text)
	}

	return sb.String()
}

func (b *block) plainText() string {
	switch b.kind {
	case blockHeading, blockCode:
		return b.text

	case blockParagraph, blockQuote, blockImage:
		return spanText(b.spans)

	case blockList:
		var lines []string
		for _, item := range b.items {
			lines = append(lines, spanText(item.spans))
		}

		return strings.Join(lines, "\n")

	case blockTable:
		var lines []string
		for _, row := range b.rows {
			lines = append(lines, strings.Join(row, " "))
		}

		return strings.Join(lines, "\n")

	case blockDivider:
		return ""

	default:
		return ""
	}
}
