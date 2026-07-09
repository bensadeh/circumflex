// Package meta renders the story meta block: the framed box a detail view
// draws above its content. Each block variant lives in its own file and owns
// its layout; this file holds the pieces they share — the frame, the column
// grid, and the Block type. A variant's Skeleton derives from the same body
// as its Render, so redesigning a block can never leave its loading stand-in
// a different size.
package meta

import (
	"strings"

	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
)

const (
	borderSize  = 2
	paddingSize = 2
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
	NewComments   int
	RootComment   string // story self-text, already rendered and wrapped by the caller
	NerdFonts     bool
}

// Block is one meta block variant bound to its data. Render draws the loaded
// block; Skeleton draws the loading stand-in: an empty, dimmed frame with
// exactly the dimensions Render produces from the same Data, so the box
// neither moves nor resizes when the content fills it in. width is the text
// column the block sits above (the comment column or the article column).
type Block struct {
	body func(width int) string
}

func (b Block) Render(width int) string {
	return frame(width).Render(b.body(width))
}

func (b Block) Skeleton(width int) string {
	// The row count comes from the framed render, not the raw body: a body
	// line wider than the frame wraps when drawn, and the skeleton must grow
	// with it.
	rows := lipgloss.Height(b.Render(width)) - borderSize

	lines := strings.Split(frame(width).Render(strings.Repeat("\n", rows-1)), "\n")
	for i, line := range lines {
		lines[i] = style.Faint(line)
	}

	return strings.Join(lines, "\n")
}

func frame(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		PaddingRight(1).
		MarginLeft(1).
		Width(width + borderSize)
}

// columns lays two texts out side by side, the left flushed left and the
// right flushed right, each taking half the content width.
func columns(contentWidth int, left, right string) string {
	columnWidth := contentWidth / 2

	l := lipgloss.NewStyle().Width(columnWidth).Align(lipgloss.Left).Render(left)
	r := lipgloss.NewStyle().Width(columnWidth).Align(lipgloss.Right).Render(right)

	return lipgloss.JoinHorizontal(lipgloss.Left, l, r)
}
