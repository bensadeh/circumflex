package article

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/syntax"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

const (
	sectionMarker = "■"
	imageCircle   = "●"
	blockIndent   = "  "
)

// Prose wraps at the reading column; code and verbatim blocks break out to
// codeWidth (the full screen space), mirroring the comment section.
func renderBlocks(blocks []block, width, codeWidth int) string {
	var parts []string

	for i := range blocks {
		if rendered := renderBlock(&blocks[i], width, codeWidth); rendered != "" {
			parts = append(parts, rendered)
		}
	}

	return strings.Join(parts, "\n\n")
}

func renderBlock(b *block, width, codeWidth int) string {
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
		return renderCode(b.text, codeWidth)

	case blockTable:
		return renderTable(b.rows, width)

	case blockImage:
		return renderImage(b.spans, width)

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

		case formatItalic:
			// Quotes are rendered in italics, so italic runs invert instead.
			if insideQuote {
				rendered = ansi.ItalicOff + s.text + ansi.Italic
			} else {
				rendered = ansi.Italic + s.text + ansi.ItalicOff
			}

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
			rendered = lipgloss.NewStyle().Hyperlink(s.href).Render(rendered)
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

func renderCode(text string, width int) string {
	wrapped := lipgloss.Wrap(text, width-len(blockIndent), "")

	return style.PrefixLines(styleLines(wrapped, style.Faint), blockIndent)
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

func renderImage(caption []span, width int) string {
	text := imageLabel() + spanText(caption) + ansi.Reset
	wrapped := lipgloss.Wrap(text, width-len(blockIndent), "")

	return style.PrefixLines(wrapped, blockIndent)
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
