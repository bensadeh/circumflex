package header

import (
	"image/color"
	"strings"
	"sync/atomic"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// memorialActive mirrors the HN front-page black bar. It is global because the
// status is a single fact about HN, not per-view state, and every view's header
// underline must reflect it. Set once when the fetch result arrives.
var memorialActive atomic.Bool

func SetMemorial(active bool) { memorialActive.Store(active) }

func MemorialActive() bool { return memorialActive.Load() }

// Underline renders the header rule, tinted muted gray when the HN memorial
// black bar is up. Every view draws its top rule through here so the indicator
// stays consistent across the list, help, comment, and reader screens.
func Underline(width int) string {
	if memorialActive.Load() {
		return style.MemorialUnderline(strings.Repeat("═", width))
	}

	return strings.Repeat("‾", width)
}

// Header renders the logo (or spinner) and category tabs. With dim set, the
// logo colors render faint, matching a list that is dimmed behind an open
// story in the wide layout.
func Header(allCategories []categories.Category, selectedSubHeader int, width int, spinnerView string, dim bool) string {
	return assemble(logo(spinnerView, dim), getCategories(allCategories, selectedSubHeader), width)
}

// OptionGroup is one search filter rendered as a segmented control: every
// option visible, the active one colored like a selected tab.
type OptionGroup struct {
	Options []string
	Active  int
}

// searchGroupMinGap keeps the sort and date groups apart when the pane is
// too narrow for the right alignment to provide the separation.
const searchGroupMinGap = 3

// SearchHeader renders the header in search mode: the tab row makes way for
// the filters, each group styled exactly like the category tabs — cycling a
// filter moves its highlight instead of swapping a word. The groups separate
// by alignment rather than glyphs: the last group's right edge sits at
// rightEdge, the column the front-page help panels end on.
func SearchHeader(groups []OptionGroup, rightEdge, width int, spinnerView string, dim bool) string {
	rendered := make([]string, 0, len(groups))
	for groupIdx, group := range groups {
		rendered = append(rendered, renderGroup(group, groupIdx))
	}

	content := strings.Join(rendered, strings.Repeat(" ", searchGroupMinGap))

	if len(rendered) > 1 {
		title := logo(spinnerView, dim)
		left := strings.Join(rendered[:len(rendered)-1], strings.Repeat(" ", searchGroupMinGap))
		right := rendered[len(rendered)-1]

		gap := max(searchGroupMinGap,
			rightEdge-lipgloss.Width(title)-lipgloss.Width(left)-lipgloss.Width(right))

		content = left + strings.Repeat(" ", gap) + right
	}

	return assemble(logo(spinnerView, dim), content, width)
}

func renderGroup(group OptionGroup, groupIdx int) string {
	separator := lipgloss.NewStyle().Faint(true).Render(" • ")

	var out strings.Builder

	for i, option := range group.Options {
		if i > 0 {
			out.WriteString(separator)
		}

		active := i == group.Active
		out.WriteString(lipgloss.NewStyle().
			Foreground(activeOptionColor(groupIdx, active)).
			Faint(!active).
			Render(option))
	}

	return out.String()
}

// activeOptionColor colors a group's active option from the same palette as
// the selected category tab, one color per group.
func activeOptionColor(groupIdx int, active bool) color.Color {
	if !active {
		return lipgloss.NoColor{}
	}

	return getSelectedCategoryColor(groupIdx)
}

func logo(spinnerView string, dim bool) string {
	leftPad := strings.Repeat(" ", layout.HeaderLogoLeftPadding)
	rightPad := strings.Repeat(" ", layout.HeaderLogoRightPadding)

	switch {
	case spinnerView != "":
		return spinnerView
	case memorialActive.Load():
		return leftPad + style.Faint("clx") + rightPad
	case dim:
		return leftPad + style.LogoFaint("c", "l", "x") + rightPad
	default:
		return leftPad + style.Logo("c", "l", "x") + rightPad
	}
}

func assemble(title, content string, width int) string {
	filler := getFiller(title, content, width)
	row := xansi.Truncate(title+content+filler, width, "")

	return row + "\n" + Underline(width)
}

func HelpHeader(title string, width int) string {
	padded := strings.Repeat(" ", layout.HeaderLeftMargin) + lipgloss.NewStyle().Bold(true).Render(title)

	return xansi.Truncate(padded, width, "") + "\n" + Underline(width)
}

func getFiller(title string, categories string, width int) string {
	availableSpace := width - lipgloss.Width(title+categories)

	if availableSpace < 0 {
		return ""
	}

	return strings.Repeat(" ", availableSpace)
}

func getCategories(allCategories []categories.Category, selectedSubHeader int) string {
	cats := allCategories[1:]

	var out strings.Builder

	separator := lipgloss.NewStyle().
		Faint(true).
		Render(" • ")

	for i, cat := range cats {
		name := categories.Name(cat)
		categoryColor, isSelected := getColor(i+1, selectedSubHeader)

		out.WriteString(lipgloss.NewStyle().
			Foreground(categoryColor).
			Faint(!isSelected).
			Render(name))

		if i < len(cats)-1 {
			out.WriteString(separator)
		}
	}

	return out.String()
}

func getColor(index int, selectedSubHeader int) (clr color.Color, isSelected bool) {
	if index == selectedSubHeader {
		return getSelectedCategoryColor(selectedSubHeader), true
	}

	return lipgloss.NoColor{}, false
}

func getSelectedCategoryColor(selectedSubHeader int) color.Color {
	primary := style.HeaderPrimary()
	secondary := style.HeaderSecondary()
	tertiary := style.HeaderTertiary()

	switch selectedSubHeader % 3 {
	case 0:
		return tertiary
	case 1:
		return primary
	default:
		return secondary
	}
}
