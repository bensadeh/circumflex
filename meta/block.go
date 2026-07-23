// Package meta renders the story meta block: the header a detail view draws
// above its content. Each block variant lives in its own file and owns its
// layout; this file holds the Block type, and the box around a block comes
// from the frame package — the same frame the help screens' panels use, so
// the two can't drift apart. A variant's Skeleton derives from the same body
// as its Render, so redesigning a block can never leave its loading stand-in
// a different shape.
package meta

import (
	"strings"

	"github.com/bensadeh/circumflex/frame"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
)

// Data is the story metadata a block can draw. A variant reads only the
// fields it shows; leave the rest zero. Zero-valued strings mean "unknown"
// and render nothing that depends on them.
type Data struct {
	URL           string
	Domain        string
	Author        string
	TimeAgo       string
	ID            int
	Points        int
	CommentsCount int
	NewComments   int    // comments since the last visit; 0 when unknown
	RootComment   string // story self-text, already rendered and wrapped by the caller
	NerdFonts     bool
}

// Block is one meta block variant bound to its data. Render draws the loaded
// block; Skeleton draws the loading stand-in: the same frame around the same
// number of rows, with the text yet to fill in, so nothing moves when the
// content arrives. width is the text column the block sits above (the
// comment column or the article column).
//
// The output carries no left margin. The hosting view indents the block with
// the same margin it gives the column's text — one margin, applied in one
// place, is what keeps the block flush with the text below it.
type Block struct {
	title         string   // sits in the frame's opening rule like a help-panel title
	labels        []string // right-aligned group closing the opening rule; sheds from the left when narrow
	closingLabels []string // left-aligned group opening the bottom rule (the story id); empty leaves it plain
	body          func(width int) string
}

// ContentWidth is the width of the text inside a block laid out at width;
// callers wrap content they pre-render themselves (the root comment) at this
// width.
func ContentWidth(width int) int {
	return frame.ContentWidth(width)
}

func (b Block) Render(width int) string {
	rows := []string{frame.OpeningRule(b.title, b.labels, width)}

	if body := b.body(width); body != "" {
		for line := range strings.SplitSeq(body, "\n") {
			rows = append(rows, frame.Row(line, width))
		}
	}

	return frame.Join(append(rows, frame.ClosingRule(b.closingLabels, width)), width)
}

func (b Block) Skeleton(width int) string {
	// The row count comes from the render so the skeleton grows with
	// whatever the body wraps to.
	rows := lipgloss.Height(b.Render(width))

	lines := make([]string, rows)
	lines[0] = frame.OpeningRule("", nil, width)

	for i := 1; i < rows-1; i++ {
		lines[i] = frame.Row("", width)
	}

	lines[rows-1] = frame.ClosingRule(nil, width)

	return frame.Join(lines, width)
}

// divider is a faint rule across the content width, drawn where two kinds
// of prose inside a block would otherwise blur together. Panes smaller than
// the block's insets leave no content column at all; the rule just vanishes
// with it.
func divider(contentWidth int) string {
	return style.Faint(strings.Repeat("─", max(0, contentWidth)))
}
