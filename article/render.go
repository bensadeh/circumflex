package article

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/syntax"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

const (
	sectionMarker     = "■"
	imageCircle       = "●"
	blockIndent       = "  "
	maxImageRows      = 40     // cap rendered image height so a tall image can still scroll past
	minImageCols      = 8      // floor so a scaled-down thumbnail stays visible
	referenceColumnPx = 640    // display width in CSS px that maps to the full content column
	minAlpha          = 0x8000 // without a known terminal background, below half coverage renders transparent
)

// ImageOptions controls how image blocks render.
type ImageOptions struct {
	// Show renders decoded images as half-block art instead of a text label.
	Show bool
	// TerminalBG is the terminal's background color, used to composite
	// semi-transparent pixels. When nil, pixels below half coverage render
	// fully transparent instead of blended.
	TerminalBG color.Color
}

// Prose wraps at the reading column. Code renders in a box spanning at least
// that column, growing with long lines up to codeWidth (the space left of
// the scrollbar); verbatim and table blocks break out to codeWidth directly.
func renderBlocks(blocks []block, width, codeWidth int, images ImageOptions) string {
	return strings.Join(renderParts(blocks, width, codeWidth, images), "\n\n")
}

// renderParts renders each block to its own chunk, skipping blocks that
// render empty. Chunks are joined with one blank line between them.
func renderParts(blocks []block, width, codeWidth int, images ImageOptions) []string {
	var parts []string

	for i := range blocks {
		if rendered := renderBlock(&blocks[i], width, codeWidth, images); rendered != "" {
			parts = append(parts, rendered)
		}
	}

	return parts
}

// blockStarts returns the line index each part lands on in the joined output,
// with the first part starting at firstLine. Scroll positions can then be
// re-anchored to the same block across re-renders.
func blockStarts(parts []string, firstLine int) []int {
	starts := make([]int, len(parts))
	line := firstLine

	for i, part := range parts {
		starts[i] = line
		line += strings.Count(part, "\n") + 2 // the part's lines plus the blank separator
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
		return renderCode(b.text, width, codeWidth)

	case blockTable:
		return renderTable(b.rows, codeWidth)

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
			if insideQuote {
				rendered = s.text
			} else {
				rendered = ansi.Reset + style.CommentBacktick(s.text)
			}

		case formatStrike:
			rendered = ansi.Strikethrough + s.text + ansi.StrikethroughOff

		default:
			rendered = s.text
		}

		if s.href != "" {
			rendered = style.ReaderLink(rendered, s.href)
		}

		sb.WriteString(rendered)
	}

	return sb.String()
}

func renderParagraph(spans []span, width int) string {
	text := renderSpans(spans, false)
	text = syntax.HighlightMentions(text)

	return lipgloss.Wrap(text, width, "")
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
func renderCode(text string, width, codeWidth int) string {
	wrapped := lipgloss.Wrap(text, codeWidth-style.RoundedBoxChrome, "")

	return style.RoundedBox(styleLines(wrapped, style.Faint), width)
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
	if images.Show && b.img != nil {
		if part := cachedImagePart(b, width, images.TerminalBG); part != "" {
			return part
		}
	}

	caption := spanText(b.spans)

	// A bare label for an image that was deliberately skipped as decoration
	// (divider strips, tracking pixels) tells the reader nothing.
	if b.decorative && caption == "" {
		return ""
	}

	inner := width - len(blockIndent)
	text := imageLabel() + caption + ansi.Reset
	wrapped := lipgloss.Wrap(text, inner, "")

	return style.PrefixLines(wrapped, blockIndent)
}

// artKey identifies the inputs the rendered art depends on, so a cached part
// survives image toggling but not a resize or background change.
type artKey struct {
	width   int // never 0 for a real render, so the zero key means "not cached"
	bg      color.RGBA
	bgKnown bool
}

// cachedImagePart renders an image block's half-block art with its centered
// caption, memoized on the block: hiding and re-showing images (or scrolling
// re-renders) reuse it instead of re-sampling every pixel.
func cachedImagePart(b *block, width int, bg color.Color) string {
	key := artKey{width: width}
	if bg != nil {
		key.bg, _ = color.RGBAModel.Convert(bg).(color.RGBA)
		key.bgKnown = true
	}

	if b.artFor == key {
		return b.art
	}

	inner := width - len(blockIndent)

	part := ""

	if art := renderImageArt(b.img, b.dispWidth, inner, bg); art != "" {
		art = centerLines(art, inner)

		if caption := spanText(b.spans); caption != "" {
			art += "\n" + centerLines(captionLines(caption, inner), inner)
		}

		part = style.PrefixLines(art, blockIndent)
	}

	b.art, b.artFor = part, key

	return part
}

// renderImageArt downsamples img and prints it with the upper half-block ▀:
// the glyph's foreground color is the top pixel and the cell's background color
// is the pixel below it, so one text row shows two pixel rows. The width tracks
// how large the image appears on the page (dispWidth, or its intrinsic size
// when unknown) relative to a reference column, so a thumbnail stays a
// thumbnail rather than filling availCols.
func renderImageArt(img image.Image, dispWidth, availCols int, bg color.Color) string {
	bounds := img.Bounds()
	if availCols < 1 || bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return ""
	}

	gridW := imageCols(dispWidth, bounds.Dx(), availCols)

	gridH := max(2, gridW*bounds.Dy()/bounds.Dx())

	if maxH := maxImageRows * 2; gridH > maxH {
		gridH = maxH
		gridW = max(1, gridH*bounds.Dx()/bounds.Dy())
	}

	if gridH%2 == 1 {
		gridH++
	}

	var sb strings.Builder

	rows := gridH / 2

	for row := range rows {
		for col := range gridW {
			top := samplePixel(img, bounds, col, 2*row, gridW, gridH, bg)
			bottom := samplePixel(img, bounds, col, 2*row+1, gridW, gridH, bg)
			writeCell(&sb, top, bottom)
		}

		sb.WriteString(ansi.Reset)

		if row < rows-1 {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// rgb8 holds the decimal strings for 0-255: a screenful of art writes
// hundreds of thousands of cells, too hot for fmt.
var rgb8 = func() (s [256]string) {
	for i := range s {
		s[i] = strconv.Itoa(i)
	}

	return s
}()

// writeCell prints one terminal cell covering two pixels. Transparent pixels
// keep the terminal's own background (a logo cut-out shows the terminal, not a
// guessed page color), so a half-covered cell uses the half-block that leaves
// the transparent side unpainted.
func writeCell(sb *strings.Builder, top, bottom pixel) {
	switch {
	case top.opaque && bottom.opaque:
		sb.WriteString("\x1b[38;2;")
		writeRGB(sb, top)
		sb.WriteString(";48;2;")
		writeRGB(sb, bottom)
		sb.WriteString("m▀")

	case top.opaque:
		sb.WriteString("\x1b[49;38;2;")
		writeRGB(sb, top)
		sb.WriteString("m▀")

	case bottom.opaque:
		sb.WriteString("\x1b[49;38;2;")
		writeRGB(sb, bottom)
		sb.WriteString("m▄")

	default:
		sb.WriteString("\x1b[49m ")
	}
}

func writeRGB(sb *strings.Builder, p pixel) {
	sb.WriteString(rgb8[p.r])
	sb.WriteByte(';')
	sb.WriteString(rgb8[p.g])
	sb.WriteByte(';')
	sb.WriteString(rgb8[p.b])
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

type pixel struct {
	r, g, b uint8
	opaque  bool
}

// samplePixel nearest-neighbour samples the source pixel for grid cell
// (gx, gy). Fully transparent pixels stay unpainted so the terminal shows
// through. Semi-transparent ones composite onto the terminal's background
// when it is known; without it, pixels below half coverage count as
// transparent and the rest are un-premultiplied so a semi-transparent edge
// keeps its own hue instead of darkening toward black.
func samplePixel(img image.Image, bounds image.Rectangle, gx, gy, gridW, gridH int, bg color.Color) pixel {
	sx := bounds.Min.X + gx*bounds.Dx()/gridW
	sy := bounds.Min.Y + gy*bounds.Dy()/gridH

	var r, g, b, a uint32

	// boundImage returns *image.RGBA, so most samples take the fast path
	// instead of boxing a color.Color per pixel.
	if rgba, ok := img.(*image.RGBA); ok {
		c := rgba.RGBAAt(sx, sy)
		r, g, b, a = uint32(c.R)*0x101, uint32(c.G)*0x101, uint32(c.B)*0x101, uint32(c.A)*0x101
	} else {
		r, g, b, a = img.At(sx, sy).RGBA()
	}

	switch {
	case a == 0xffff:
		return pixel{r: uint8(r >> 8), g: uint8(g >> 8), b: uint8(b >> 8), opaque: true}

	case bg != nil:
		if a == 0 {
			return pixel{}
		}

		return compositePixel(r, g, b, a, bg)

	case a < minAlpha:
		return pixel{}

	default:
		return pixel{
			r:      uint8(r * 0xffff / a >> 8),
			g:      uint8(g * 0xffff / a >> 8),
			b:      uint8(b * 0xffff / a >> 8),
			opaque: true,
		}
	}
}

// compositePixel blends a premultiplied pixel onto the terminal background —
// the same math a browser uses to draw the image over the page.
func compositePixel(r, g, b, a uint32, bg color.Color) pixel {
	bgR, bgG, bgB, _ := bg.RGBA()
	inv := 0xffff - a

	return pixel{
		r:      uint8((r + inv*bgR/0xffff) >> 8),
		g:      uint8((g + inv*bgG/0xffff) >> 8),
		b:      uint8((b + inv*bgB/0xffff) >> 8),
		opaque: true,
	}
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
	circles := lipgloss.NewStyle().Foreground(style.HeaderC()).Faint(true).Render(imageCircle) +
		lipgloss.NewStyle().Foreground(style.HeaderL()).Faint(true).Render(imageCircle) +
		lipgloss.NewStyle().Foreground(style.HeaderX()).Faint(true).Render(imageCircle)

	title := lipgloss.NewStyle().Foreground(style.ReaderImageColor()).Faint(true).Italic(true).Render(" Image ")

	return ansi.Reset + circles + ansi.Reset + title + ansi.Faint + ansi.Italic
}

func renderTable(rows [][]string, width int) string {
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

		if rowIndex == 0 && len(rows) > 1 {
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
