package article

import (
	"fmt"
	"image"
	"regexp"
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/highlight"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
)

const (
	sectionMarker     = "■"
	imageCircle       = "●"
	blockIndent       = "  "
	maxImageRows      = 40  // cap rendered image height so a tall image can still scroll past
	minImageCols      = 8   // floor so a scaled-down thumbnail stays visible
	referenceColumnPx = 640 // display width in CSS px that maps to the full content column
)

// ImageOptions controls how image blocks render.
type ImageOptions struct {
	// Show renders image blocks as pixels instead of a text label.
	Show bool
	// Kitty reports that the terminal speaks the Kitty graphics protocol.
	// Without it there is no way to draw an image, so every image block
	// renders as its label regardless of Show.
	Kitty bool
	// CellWidth and CellHeight are one terminal cell's pixel dimensions,
	// which keep an image's aspect ratio honest in cells. Zero assumes cells
	// twice as tall as wide.
	CellWidth, CellHeight int
}

// renderedPart is one block's rendered chunk plus the kind that produced it,
// so line positions in the joined output can be attributed to structure.
type renderedPart struct {
	text string
	kind blockKind
}

// Prose wraps at the reading column. Code renders in a box spanning at least
// that column, growing with long lines up to codeWidth (the space left of
// the scrollbar); verbatim and table blocks break out to codeWidth directly.
func renderBlocks(blocks []block, width, codeWidth int, images ImageOptions) string {
	return joinParts(renderParts(blocks, width, codeWidth, images))
}

// renderParts renders each block to its own chunk, skipping blocks that
// render empty. Chunks are joined with one blank line between them.
func renderParts(blocks []block, width, codeWidth int, images ImageOptions) []renderedPart {
	var parts []renderedPart

	for i := range blocks {
		if rendered := renderBlock(&blocks[i], width, codeWidth, images); rendered != "" {
			parts = append(parts, renderedPart{text: rendered, kind: blocks[i].kind})
		}
	}

	return parts
}

// partSeparator is the blank line between rendered blocks. blockStarts
// derives its line offsets from the same constant, so the joined output and
// the block positions cannot drift apart.
const partSeparator = "\n\n"

func joinParts(parts []renderedPart) string {
	texts := make([]string, len(parts))
	for i, part := range parts {
		texts[i] = part.text
	}

	return strings.Join(texts, partSeparator)
}

// blockStarts returns the line index each part lands on in the joined output,
// with the first part starting at firstLine. Scroll positions can then be
// re-anchored to the same block across re-renders.
func blockStarts(parts []renderedPart, firstLine int) []int {
	starts := make([]int, len(parts))
	line := firstLine

	for i, part := range parts {
		starts[i] = line
		line += strings.Count(part.text+partSeparator, "\n")
	}

	return starts
}

func renderBlock(b *block, width, codeWidth int, images ImageOptions) string {
	switch b.kind {
	case blockParagraph:
		return renderParagraph(b.spans, width)

	case blockHeading:
		return renderHeading(b.level, b.text, width)

	case blockList:
		return renderList(b.items, width)

	case blockQuote:
		return renderQuote(b.spans, width)

	case blockCode:
		return renderCode(b, width, codeWidth)

	case blockTable:
		return renderTable(b.rows, b.hasHeader, codeWidth)

	case blockImage:
		return renderImage(b, width, images)

	case blockDivider:
		return renderDivider(width)

	case blockVerbatim:
		return lipgloss.Wrap(b.text, codeWidth, "")

	default:
		return ""
	}
}

func renderSpans(spans []span, insideQuote bool) string {
	var sb strings.Builder

	for _, s := range spans {
		var rendered string

		// Links reader mode can't open in place carry a dashed underline,
		// matching the inert marking the URL selector gives them.
		viewable := s.href != "" && ValidateURL(s.href) == nil

		switch s.format {
		case formatPlain:
			rendered = s.text

		case formatBold:
			// NormalIntensity clears faint along with bold, so the quote's
			// faint is re-opened.
			rendered = ansi.Bold + s.text + ansi.NormalIntensity
			if insideQuote {
				rendered += ansi.Faint
			}

		case formatItalic:
			// Quotes are rendered in italics, so italic runs invert instead.
			if insideQuote {
				rendered = ansi.ItalicOff + s.text + ansi.Italic
			} else {
				rendered = ansi.Italic + s.text + ansi.ItalicOff
			}

		case formatUnderline:
			rendered = ansi.Underline + s.text + ansi.UnderlineOff

		case formatCode:
			switch {
			case insideQuote:
				rendered = s.text
			case s.href != "" && viewable:
				// The leading reset clears the link wrapper's underline along
				// with everything else, so the backtick style re-adds it.
				rendered = ansi.Reset + style.CommentBacktickLink(s.text)
			case s.href != "":
				rendered = ansi.Reset + style.CommentBacktickLinkInert(s.text)
			default:
				rendered = ansi.Reset + style.CommentBacktick(s.text)
			}

		case formatStrike:
			rendered = ansi.Strikethrough + s.text + ansi.StrikethroughOff

		default:
			rendered = s.text
		}

		switch {
		case viewable:
			rendered = style.ReaderLink(rendered, s.href)
		case s.href != "":
			rendered = style.ReaderLinkInert(rendered, s.href)
		}

		sb.WriteString(rendered)
	}

	return sb.String()
}

func renderParagraph(spans []span, width int) string {
	text := renderSpans(spans, false)
	text = highlightMentions(text)

	return lipgloss.Wrap(text, width, "")
}

var reMention = regexp.MustCompile(`((?:^| )\B@[\w.]+)`)

// highlightMentions colors @handles in article prose, giving @dang the mod
// color. HN discussions embedded in articles read like comments.
func highlightMentions(input string) string {
	input = reMention.ReplaceAllString(input, style.CommentMention(`$1`))

	input = strings.ReplaceAll(input, style.CommentMention("@dang"),
		style.CommentMod("@dang"))
	input = strings.ReplaceAll(input, style.CommentMention(" @dang"),
		style.CommentMod(" @dang"))

	return input
}

func renderHeading(level int, text string, width int) string {
	indent := (level - 1) * 2

	styled := headingStyle(level)(sectionMarker+" ") + style.Bold(text)
	wrapped := lipgloss.Wrap(styled, width-indent, "")

	if indent == 0 {
		return wrapped
	}

	return style.PrefixLines(wrapped, strings.Repeat(" ", indent))
}

func headingStyle(level int) func(string) string {
	switch level {
	case 1:
		return style.ReaderH1
	case 2:
		return style.ReaderH2
	case 3:
		return style.ReaderH3
	case 4:
		return style.ReaderH4
	case 5:
		return style.ReaderH5
	default:
		return style.ReaderH6
	}
}

func listGlyph(depth int) string {
	switch depth {
	case 0:
		return "-"
	case 1:
		return "•"
	case 2:
		return "◦"
	case 3:
		return "▪"
	default:
		return "▫"
	}
}

func renderList(items []listItem, width int) string {
	numberWidth := 0
	for _, item := range items {
		numberWidth = max(numberWidth, len(strconv.Itoa(item.number)))
	}

	var lines []string

	for _, item := range items {
		token := listGlyph(item.depth)
		if item.number > 0 {
			token = fmt.Sprintf("%*d.", numberWidth, item.number)
		}

		head := strings.Repeat(blockIndent, item.depth+1) + token + " "
		continuation := strings.Repeat(" ", lipgloss.Width(head))

		wrapped := lipgloss.Wrap(renderSpans(item.spans, false), width-lipgloss.Width(head), "")

		for i, line := range strings.Split(wrapped, "\n") {
			if i == 0 {
				lines = append(lines, head+line)
			} else {
				lines = append(lines, continuation+line)
			}
		}
	}

	return strings.Join(lines, "\n")
}

func renderQuote(spans []span, width int) string {
	prefix := blockIndent + style.Faint(" "+style.IndentSymbol)

	quoteStyle := lipgloss.NewStyle().Italic(true).Faint(true)
	wrapped := lipgloss.Wrap(renderSpans(spans, true), width-lipgloss.Width(prefix), "")

	styled := styleLines(wrapped, func(line string) string { return quoteStyle.Render(line) })

	return style.PrefixLines(styled, prefix)
}

// The box spans at least the reading column and grows with long code lines
// up to codeWidth; its border and padding equal the old blockIndent, so the
// code text keeps its column.
func renderCode(b *block, width, codeWidth int) string {
	// Tokenizing is width-independent and costs real time on big blocks, so
	// it runs once per block, not once per resize step.
	if !b.hlDone {
		b.hlOut = highlight.Code(b.text, b.lang)
		b.hlDone = true
	}

	if b.hlOut != "" {
		return highlight.Boxed(b.hlOut, b.lang, codeWidth, width)
	}

	wrapped := style.WrapWithin(b.text, codeWidth-style.RoundedBoxChrome)

	return style.RoundedBox(styleLines(wrapped, style.Faint), width, "")
}

// Styling line by line, because lipgloss pads multi-line strings to a uniform
// width, leaving trailing whitespace on every line.
func styleLines(text string, styleFn func(string) string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = styleFn(line)
	}

	return strings.Join(lines, "\n")
}

func renderImage(b *block, width int, images ImageOptions) string {
	if images.Kitty && b.kitty != nil && b.imgSize != (image.Point{}) {
		if images.Show {
			if part := cachedImagePart(b, width, images); part != "" {
				return part
			}
		} else {
			// A hidden image still records the grid a show at this width would
			// lay down, so its pixels reach the terminal while it is hidden and
			// the first show composites as instantly as every later one.
			recordKittyGrid(b, width-2*len(blockIndent), images.CellWidth, images.CellHeight)
		}
	}

	// An image skipped as decoration (badges, divider strips, tracking
	// pixels) stays dropped even when captioned: a badge's alt text is
	// chrome, not content.
	if b.decorative {
		return ""
	}

	caption := spanText(b.spans)

	label := imageLabel()
	if b.figure {
		label = figureLabel()
	}

	inner := width - len(blockIndent)
	text := label + caption + ansi.Reset
	wrapped := lipgloss.Wrap(text, inner, "")

	return style.PrefixLines(wrapped, blockIndent)
}

// artKey identifies the inputs the rendered art depends on, so a cached part
// survives image toggling but not a resize or a font-size change.
type artKey struct {
	width int // never 0 for a real render, so the zero key means "not cached"
	cellW int
	cellH int
}

// cachedImagePart renders an image block's placeholder cells with its
// centered caption, memoized on the block: hiding and re-showing images (or
// scrolling re-renders) reuse it instead of re-deriving the grid.
func cachedImagePart(b *block, width int, images ImageOptions) string {
	key := artKey{width: width, cellW: images.CellWidth, cellH: images.CellHeight}
	if b.artFor == key {
		return b.art
	}

	// The art keeps the block indent on both sides, so a full-width image
	// stops short of the right edge like the left.
	inner := width - 2*len(blockIndent)

	art := renderKittyArt(b, inner, key.cellW, key.cellH)
	part := ""

	if art != "" {
		art = centerLines(art, inner)

		if caption := spanText(b.spans); caption != "" {
			art += "\n" + centerLines(captionLines(caption, inner), inner)
		}

		part = style.PrefixLines(art, blockIndent)
	}

	b.art, b.artFor = part, key

	return part
}

// renderKittyArt lays down the cell grid a transmitted image composites
// onto: rows of placeholder characters whose diacritics address the image's
// tile grid and whose indexed foreground color names the image. The cells
// are ordinary styled text — they wrap, scroll and diff like any other line,
// and rows scrolled off screen simply aren't drawn. Columns after the first
// omit the diacritics; the terminal infers them from the run.
func renderKittyArt(b *block, availCols, cellW, cellH int) string {
	cols, rows, ok := recordKittyGrid(b, availCols, cellW, cellH)
	if !ok {
		return ""
	}

	fg := "\x1b[38;5;" + strconv.Itoa(b.kitty.id) + "m"

	var sb strings.Builder

	for row := range rows {
		sb.WriteString(fg)
		sb.WriteRune(kitty.Placeholder)
		sb.WriteRune(kitty.Diacritic(row))
		sb.WriteRune(kitty.Diacritic(0))

		for range cols - 1 {
			sb.WriteRune(kitty.Placeholder)
		}

		sb.WriteString(ansi.Reset)

		if row < rows-1 {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// recordKittyGrid sizes b's cell grid and records it as the geometry this
// render wants, so PendingKittyWork can settle the terminal against it —
// whether the grid was laid down as placeholder cells or belongs to a hidden
// image whose pixels travel ahead of its first show. Degenerate geometry
// records nothing.
func recordKittyGrid(b *block, availCols, cellW, cellH int) (cols, rows int, ok bool) {
	if availCols < 1 || b.imgSize.X <= 0 || b.imgSize.Y <= 0 {
		return 0, 0, false
	}

	cols, rows = kittyGrid(b.dispWidth, b.imgSize.X, b.imgSize.Y, availCols, cellW, cellH)
	b.kitty.wantCols, b.kitty.wantRows = cols, rows

	return cols, rows, true
}

// kittyGrid sizes an image's cell grid: columns from its on-page size, rows
// from the image's aspect ratio scaled by the cell's pixel shape. The terminal
// stretches the image to fill the grid exactly, so the rows must account for
// cells being taller than wide — 1:2 when the terminal never reported its cell
// size.
func kittyGrid(dispWidth, imgW, imgH, availCols, cellW, cellH int) (cols, rows int) {
	cols = imageCols(dispWidth, imgW, availCols)

	if cellW <= 0 || cellH <= 0 {
		cellW, cellH = 1, 2
	}

	rows = max(1, (cols*cellW*imgH+imgW*cellH/2)/(imgW*cellH))

	if rows > maxImageRows {
		rows = maxImageRows
		cols = min(availCols, max(1, rows*cellH*imgW/(imgH*cellW)))
	}

	return cols, rows
}

// imageCols maps an image's on-page size to a terminal column count: its
// display width in CSS px (falling back to the intrinsic width) as a fraction
// of referenceColumnPx, floored so it stays visible and capped at availCols.
func imageCols(dispWidth, intrinsicWidth, availCols int) int {
	px := dispWidth
	if px <= 0 {
		px = intrinsicWidth
	}

	cols := availCols * px / referenceColumnPx

	return min(availCols, max(minImageCols, cols))
}

// centerLines pads each line to sit centered within width, so a scaled-down
// image and its caption hang in the middle of the content column instead of
// hugging the left edge. Full-width lines pass through unchanged.
func centerLines(text string, width int) string {
	return styleLines(text, func(line string) string {
		pad := (width - xansi.StringWidth(line)) / 2
		if pad <= 0 {
			return line
		}

		return strings.Repeat(" ", pad) + line
	})
}

func captionLines(caption string, width int) string {
	wrapped := lipgloss.Wrap(caption, width, "")

	return styleLines(wrapped, func(line string) string {
		return lipgloss.NewStyle().Foreground(style.ReaderImageColor()).Faint(true).Italic(true).Render(line)
	})
}

func imageLabel() string {
	return graphicLabel(imageCircle, imageCircle, imageCircle, " Image ")
}

// figureLabel marks a described graphic — one drawn in markup with no bitmap
// at all, or a chart standing in for pixels this terminal cannot draw: an
// ascending mini bar chart in place of the image circles.
func figureLabel() string {
	return graphicLabel("▂", "▄", "▆", " Figure ")
}

func graphicLabel(first, second, third, title string) string {
	marks := lipgloss.NewStyle().Foreground(style.HeaderC()).Faint(true).Render(first) +
		lipgloss.NewStyle().Foreground(style.HeaderL()).Faint(true).Render(second) +
		lipgloss.NewStyle().Foreground(style.HeaderX()).Faint(true).Render(third)

	styledTitle := lipgloss.NewStyle().Foreground(style.ReaderImageColor()).Faint(true).Italic(true).Render(title)

	return ansi.Reset + marks + ansi.Reset + styledTitle + ansi.Faint + ansi.Italic
}

func renderTable(rows [][]string, hasHeader bool, width int) string {
	columns := 0
	for _, row := range rows {
		columns = max(columns, len(row))
	}

	columnWidths := make([]int, columns)

	for _, row := range rows {
		for i, cell := range row {
			columnWidths[i] = max(columnWidths[i], lipgloss.Width(cell))
		}
	}

	var lines []string

	for rowIndex, row := range rows {
		cells := make([]string, columns)

		for i := range cells {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}

			cells[i] = cell + strings.Repeat(" ", columnWidths[i]-lipgloss.Width(cell))
		}

		lines = append(lines, strings.TrimRight(strings.Join(cells, "  "), " "))

		if rowIndex == 0 && hasHeader && len(rows) > 1 {
			separators := make([]string, columns)
			for i, columnWidth := range columnWidths {
				separators[i] = strings.Repeat("-", columnWidth)
			}

			lines = append(lines, style.Faint(strings.Join(separators, "  ")))
		}
	}

	for i := range lines {
		lines[i] = xansi.Truncate(lines[i], width, "…")
	}

	return strings.Join(lines, "\n")
}

func renderDivider(width int) string {
	return blockIndent + style.Faint(strings.Repeat("-", max(1, width-2*len(blockIndent))))
}
