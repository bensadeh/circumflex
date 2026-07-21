// Package frame draws the faint rounded box that the meta block and the
// help screens' panels share: an opening rule that doubles as the header
// row, body rows between side borders, and a closing rule. The two box
// families are one design — a frame laid out at width spans exactly width
// cells — and rendering both through this single implementation is what
// keeps their edges from drifting apart.
package frame

import (
	"slices"
	"strings"

	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

const (
	padding = 1
	chrome  = 2*padding + 2 // horizontal padding + the side borders
	lead    = 2             // rule cells between a corner and the text beside it
)

// ContentWidth is the cells available to text inside a frame spanning width.
func ContentWidth(width int) int {
	return max(0, width-chrome)
}

// OpeningRule is the frame's top border, doubling as its header row: the
// pre-styled title sits in the rule the way a panel title does, and the
// label group closes the rule against the right corner, a single rule cell
// between the labels. When the rule can't carry everything, the labels shed
// from the left and the title outlasts them — the frame never gives up its
// own corners.
func OpeningRule(title string, labels []string, width int) string {
	titleCells := lead + lipgloss.Width(title) + 4 // the corners + a space each side of the title

	if title == "" || titleCells > width {
		return style.Faint("╭" + rule(width-2) + "╮")
	}

	labels = slices.DeleteFunc(slices.Clone(labels), func(l string) bool { return l == "" })

	for ; len(labels) > 0; labels = labels[1:] {
		group := strings.Join(labels, style.Faint(" "+rule(1)+" "))
		fill := width - titleCells - lipgloss.Width(group) - lead - 2

		if fill < 1 {
			continue
		}

		return style.Faint("╭"+rule(lead)+" ") + title +
			style.Faint(" "+rule(fill)+" ") + group +
			style.Faint(" "+rule(lead)+"╮")
	}

	return style.Faint("╭"+rule(lead)+" ") + title +
		style.Faint(" "+rule(width-titleCells)+"╮")
}

// ClosingRule is the frame's bottom border. A label group closes it against
// the bottom-right corner the way the opening rule's labels close the top: a
// single rule cell between labels, shedding from the left when the rule can't
// carry them, and falling back to a plain border when nothing fits — the
// frame never gives up its own corners.
func ClosingRule(labels []string, width int) string {
	labels = slices.DeleteFunc(slices.Clone(labels), func(l string) bool { return l == "" })

	for ; len(labels) > 0; labels = labels[1:] {
		group := strings.Join(labels, style.Faint(" "+rule(1)+" "))
		fill := width - lipgloss.Width(group) - lead - 4 // the corners + a space each side of the group

		if fill < lead {
			continue
		}

		return style.Faint("╰"+rule(fill)+" ") + group + style.Faint(" "+rule(lead)+"╯")
	}

	return style.Faint("╰" + rule(width-2) + "╯")
}

// Row is one row of the frame's body: the text padded to the content width
// between the side borders.
func Row(line string, width int) string {
	gutter := strings.Repeat(" ", padding)
	pad := strings.Repeat(" ", max(0, ContentWidth(width)-lipgloss.Width(line)))

	return style.Faint("│") + gutter + line + pad + gutter + style.Faint("│")
}

// Join joins the frame's rows, cutting any row wider than the frame — a
// pane narrower than the frame's own chrome sheds the box's right side
// rather than spilling into the column next door.
func Join(rows []string, width int) string {
	width = max(0, width)

	for i, row := range rows {
		if lipgloss.Width(row) > width {
			rows[i] = xansi.Truncate(row, width, "")
		}
	}

	return strings.Join(rows, "\n")
}

func rule(cells int) string {
	return strings.Repeat("─", max(0, cells))
}
