package header

import (
	"image/color"
	"strings"
	"sync/atomic"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
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
	bar := strings.Repeat("‾", width)
	if memorialActive.Load() {
		return style.MemorialUnderline(bar)
	}

	return bar
}

func Header(allCategories []categories.Category, selectedSubHeader int, width int, spinnerView string) string {
	leftPad := strings.Repeat(" ", layout.HeaderLogoLeftPadding)
	rightPad := strings.Repeat(" ", layout.HeaderLogoRightPadding)

	var title string
	if spinnerView != "" {
		title = spinnerView
	} else {
		title = leftPad + style.Logo("c", "l", "x") + rightPad
	}

	cats := getCategories(allCategories, selectedSubHeader)
	filler := getFiller(title, cats, width)

	return title + cats + filler + "\n" + Underline(width)
}

func HelpHeader(title string, width int) string {
	padded := strings.Repeat(" ", layout.HeaderLeftMargin) + lipgloss.NewStyle().Bold(true).Render(title)

	return padded + "\n" + Underline(width)
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
