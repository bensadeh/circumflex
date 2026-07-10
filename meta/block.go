// Package meta renders the story meta block: the accent-barred header a
// detail view draws above its content. Each block variant lives in its own
// file and owns its layout; this file holds the pieces they share — the
// accent bar, the label stack, and the Block type. A variant's Skeleton
// derives from the same body as its Render, so redesigning a block can never
// leave its loading stand-in a different shape.
package meta

import (
	"strings"

	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
)

// bar is the block's frame: a half-block accent bar down the left edge of
// the text column. Heavier than the ▎ the comment section uses for quotes
// and nesting, so the block reads as its own element rather than another
// indent level.
const bar = "▌"

// textIndent is how much deeper than the block's left edge its text sits:
// the accent bar and its trailing space.
const textIndent = 2

// rightInset is the cell the block's text stops short of the column's right
// edge: the block sits visibly inside the column it heads, its widest rows
// ending one cell in from where full text lines wrap. The insets exist so
// the frame is confined to the block's left edge — however the frame or the
// hosting margins change, no row extends past width-rightInset.
const rightInset = 1

// Data is the story metadata a block can draw. A variant reads only the
// fields it shows; leave the rest zero. Zero-valued strings mean "unknown"
// and render nothing that depends on them.
type Data struct {
	URL           string
	Domain        string
	Author        string
	TimeAgo       string
	Points        int
	CommentsCount int
	NewComments   int
	RootComment   string // story self-text, already rendered and wrapped by the caller
	NerdFonts     bool
}

// Block is one meta block variant bound to its data. Render draws the loaded
// block; Skeleton draws the loading stand-in: the same accent bar over the
// same number of rows, with the text yet to fill in, so nothing moves when
// the content arrives. width is the text column the block sits above (the
// comment column or the article column).
//
// The output carries no left margin. The hosting view indents the block with
// the same margin it gives the column's text — one margin, applied in one
// place, is what keeps the bar aligned with the text below it.
type Block struct {
	body func(width int) string
}

// ContentWidth is the width of the text inside a block laid out at width;
// callers wrap content they pre-render themselves (the root comment) at this
// width.
func ContentWidth(width int) int {
	return width - textIndent - rightInset
}

func (b Block) Render(width int) string {
	lines := strings.Split(b.body(width), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(style.Faint(bar)+" "+line, " ")
	}

	return strings.Join(lines, "\n")
}

func (b Block) Skeleton(width int) string {
	// The row count comes from the render so the skeleton grows with
	// whatever the body wraps to.
	rows := lipgloss.Height(b.Render(width))

	lines := make([]string, rows)
	for i := range lines {
		lines[i] = style.Faint(bar)
	}

	return strings.Join(lines, "\n")
}

// divider is a faint rule across the content width, drawn where two kinds
// of prose inside a block would otherwise blur together. Panes smaller than
// the block's insets leave no content column at all; the rule just vanishes
// with it.
func divider(contentWidth int) string {
	return style.Faint(strings.Repeat("─", max(0, contentWidth)))
}

// stack lays label rows out flush left, one per line, wrapped at the content
// width so a long byline breaks instead of spilling past the block's edge.
func stack(contentWidth int, rows ...string) string {
	return lipgloss.Wrap(strings.Join(rows, "\n"), contentWidth, "")
}
