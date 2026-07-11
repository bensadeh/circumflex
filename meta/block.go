// Package meta renders the story meta block: the header a detail view draws
// above its content. Each block variant lives in its own file and owns its
// layout; this file holds the pieces they share — the frame, the label
// stack, and the Block type. A variant's Skeleton derives from the same body
// as its Render, so redesigning a block can never leave its loading stand-in
// a different shape.
package meta

import (
	"strings"

	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// The block's frame is a faint rounded box in the same family as the help
// screens' panels: the byline sits in the opening rule the way a panel title
// does, the score closes that rule right-aligned, and the closing rule marks
// where the meta block ends and the content it heads begins.
const (
	framePadding = 1
	frameChrome  = 2*framePadding + 2 // horizontal padding + the side borders
	frameLead    = 2                  // rule cells between a corner and the text beside it
)

// rightInset is the cell the frame stops short of the column's right edge:
// the block sits visibly inside the column it heads, its frame ending one
// cell in from where full text lines wrap.
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
// block; Skeleton draws the loading stand-in: the same frame around the same
// number of rows, with the text yet to fill in, so nothing moves when the
// content arrives. width is the text column the block sits above (the
// comment column or the article column).
//
// The output carries no left margin. The hosting view indents the block with
// the same margin it gives the column's text — one margin, applied in one
// place, is what keeps the block flush with the text below it.
type Block struct {
	title string // sits in the opening rule like a help-panel title
	score string // right-aligned in the opening rule
	body  func(width int) string
}

// ContentWidth is the width of the text inside a block laid out at width;
// callers wrap content they pre-render themselves (the root comment) at this
// width.
func ContentWidth(width int) int {
	return max(0, width-rightInset-frameChrome)
}

func (b Block) Render(width int) string {
	rows := []string{openingRule(b.title, b.score, frameWidth(width))}

	if body := b.body(width); body != "" {
		for line := range strings.SplitSeq(body, "\n") {
			rows = append(rows, framed(line, ContentWidth(width)))
		}
	}

	return joinFrame(append(rows, closingRule(frameWidth(width))), frameWidth(width))
}

func (b Block) Skeleton(width int) string {
	// The row count comes from the render so the skeleton grows with
	// whatever the body wraps to.
	rows := lipgloss.Height(b.Render(width))

	lines := make([]string, rows)
	lines[0] = openingRule("", "", frameWidth(width))

	for i := 1; i < rows-1; i++ {
		lines[i] = framed("", ContentWidth(width))
	}

	lines[rows-1] = closingRule(frameWidth(width))

	return joinFrame(lines, frameWidth(width))
}

func frameWidth(width int) int {
	return max(0, width-rightInset)
}

// openingRule is the frame's top border, doubling as the block's header row:
// the title sits in the rule the way a help-panel title does, and the score
// closes the rule against the right corner. When the rule can't carry both,
// the score goes first, then the title — the frame never gives up its own
// corners.
func openingRule(title, score string, frameWidth int) string {
	titleCells := frameLead + lipgloss.Width(title) + 4 // the corners + a space each side of the title

	if title == "" || titleCells > frameWidth {
		return style.Faint("╭" + rule(frameWidth-2) + "╮")
	}

	fill := frameWidth - titleCells - lipgloss.Width(score) - frameLead - 2

	if score == "" || fill < 1 {
		return style.Faint("╭"+rule(frameLead)+" ") + title +
			style.Faint(" "+rule(frameWidth-titleCells)+"╮")
	}

	return style.Faint("╭"+rule(frameLead)+" ") + title +
		style.Faint(" "+rule(fill)+" ") + score +
		style.Faint(" "+rule(frameLead)+"╮")
}

// closingRule is the frame's bottom border, drawn as the block's last row in
// both the render and the skeleton — while a story loads, the empty frame
// alone marks the rows the block will fill.
func closingRule(frameWidth int) string {
	return style.Faint("╰" + rule(frameWidth-2) + "╯")
}

func rule(cells int) string {
	return strings.Repeat("─", max(0, cells))
}

// framed is one row of the frame's body: the text padded to the content
// width between the side borders.
func framed(line string, contentWidth int) string {
	gutter := strings.Repeat(" ", framePadding)
	pad := strings.Repeat(" ", max(0, contentWidth-lipgloss.Width(line)))

	return style.Faint("│") + gutter + line + pad + gutter + style.Faint("│")
}

// joinFrame joins the frame's rows, cutting any row wider than the frame —
// a pane narrower than the frame's own chrome sheds the box's right side
// rather than spilling into the column next door.
func joinFrame(rows []string, frameWidth int) string {
	for i, row := range rows {
		if lipgloss.Width(row) > frameWidth {
			rows[i] = xansi.Truncate(row, frameWidth, "")
		}
	}

	return strings.Join(rows, "\n")
}

// divider is a faint rule across the content width, drawn where two kinds
// of prose inside a block would otherwise blur together. Panes smaller than
// the block's insets leave no content column at all; the rule just vanishes
// with it.
func divider(contentWidth int) string {
	return style.Faint(strings.Repeat("─", max(0, contentWidth)))
}
