package article

import (
	"image"
	"strings"
)

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
	blockVerbatim
)

type block struct {
	kind       blockKind
	level      int        // blockHeading: 1-6
	spans      []span     // blockParagraph, blockQuote, blockImage (caption)
	items      []listItem // blockList
	rows       [][]string // blockTable
	hasHeader  bool       // blockTable: first row came from thead or all-th cells
	text       string     // blockHeading, blockCode
	lang       string     // blockCode: page-declared language, empty when unlabeled
	hlOut      string     // blockCode: chroma render memoized by renderCode — width-independent
	hlDone     bool
	imageURL   string      // blockImage: resolved source URL, empty if none
	imgSize    image.Point // blockImage: decoded raster's dimensions, zero until fetched or on failure
	kitty      *kittyImage // blockImage: the copy and terminal state for Kitty graphics, nil when unavailable
	decorative bool        // blockImage: fetched fine but sized like a divider or tracking pixel
	figure     bool        // blockImage: known chart or diagram — labeled as one where its pixels do not render
	dispWidth  int         // blockImage: intended display width in CSS px from the width attr, 0 if unknown
	art        string      // blockImage: rendered placeholder cells memoized for artFor; see cachedImagePart
	artFor     artKey
}

type inlineFormat int

const (
	formatPlain inlineFormat = iota
	formatBold
	formatItalic
	formatUnderline
	formatCode
	formatStrike
)

type span struct {
	text   string
	format inlineFormat
	href   string
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
	case blockHeading, blockCode, blockVerbatim:
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
