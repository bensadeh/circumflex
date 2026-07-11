// Package meta renders the story meta block: the header a detail view draws
// above its content. Each block variant lives in its own file and owns its
// layout; this file holds the pieces they share — the closing rule, the label
// stack, and the Block type. A variant's Skeleton derives from the same body
// as its Render, so redesigning a block can never leave its loading stand-in
// a different shape.
package meta

import (
	"strings"

	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
)

// separator is the block's frame: a double-line rule under the block's last
// row, marking where the meta block ends and the content it heads begins.
// Double-struck so it outranks both the light ─ rule inside the block and
// the ▁ rules between comments.
const separator = "═"

// rightInset is the cell the block's text stops short of the column's right
// edge: the block sits visibly inside the column it heads, its widest rows
// ending one cell in from where full text lines wrap. The inset exists so
// the frame stays the block's business — however the frame or the hosting
// margins change, no row extends past width-rightInset.
const rightInset = 1

// Data is the story metadata a block can draw. A variant reads only the
// fields it shows; leave the rest zero. Zero-valued strings mean "unknown"
// and render nothing that depends on them.
type Data struct {
	URL         string
	Domain      string
	Author      string
	TimeAgo     string
	Points      int
	RootComment string // story self-text, already rendered and wrapped by the caller
	NerdFonts   bool
}

// Block is one meta block variant bound to its data. Render draws the loaded
// block; Skeleton draws the loading stand-in: the same closing rule over the
// same number of rows, with the text yet to fill in, so nothing moves when
// the content arrives. width is the text column the block sits above (the
// comment column or the article column).
//
// The output carries no left margin. The hosting view indents the block with
// the same margin it gives the column's text — one margin, applied in one
// place, is what keeps the block flush with the text below it.
type Block struct {
	body func(width int) string
}

// ContentWidth is the width of the text inside a block laid out at width;
// callers wrap content they pre-render themselves (the root comment) at this
// width.
func ContentWidth(width int) int {
	return width - rightInset
}

func (b Block) Render(width int) string {
	lines := strings.Split(b.body(width), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}

	return strings.Join(append(lines, closingRule(ContentWidth(width))), "\n")
}

func (b Block) Skeleton(width int) string {
	// The row count comes from the render so the skeleton grows with
	// whatever the body wraps to.
	rows := lipgloss.Height(b.Render(width))

	lines := make([]string, rows)
	lines[rows-1] = closingRule(ContentWidth(width))

	return strings.Join(lines, "\n")
}

// closingRule is the separator drawn as the block's last row, in both the
// render and the skeleton — while a story loads, the rule alone marks the
// rows the block will fill.
func closingRule(contentWidth int) string {
	return style.Faint(strings.Repeat(separator, max(0, contentWidth)))
}

// divider is a faint rule across the content width, drawn where two kinds
// of prose inside a block would otherwise blur together. Panes smaller than
// the block's insets leave no content column at all; the rule just vanishes
// with it.
func divider(contentWidth int) string {
	return style.Faint(strings.Repeat("─", max(0, contentWidth)))
}

// columns lays two texts out side by side, the left flushed left and the
// right flushed right, splitting the content width between them. The right
// column takes the odd cell so its text always ends exactly on the block's
// right edge.
func columns(contentWidth int, left, right string) string {
	leftWidth := contentWidth / 2
	rightWidth := contentWidth - leftWidth

	l := lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Left).Render(left)
	r := lipgloss.NewStyle().Width(rightWidth).Align(lipgloss.Right).Render(right)

	return lipgloss.JoinHorizontal(lipgloss.Left, l, r)
}
