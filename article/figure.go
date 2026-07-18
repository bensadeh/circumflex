package article

import (
	"regexp"
	"strings"
)

// knownFigure reports whether an image is known to depict a chart, plot, or
// diagram — where half-block art degrades to an illegible smear, so only
// Kitty-resolution pixels render and the description carries the content
// everywhere below that tier. Only what the page declares counts,
// so photographs are never demoted: a print-style numbered caption
// ("Figure 3: …") or a description led by the graphic's genre, the alt-text
// convention for complex images ("Bar chart of …"). Format is deliberately
// not a signal — an svg is never a photo, but logos, icons and illustrations
// are svg too, and their art shows more than a label would.
func knownFigure(texts ...string) bool {
	for _, text := range texts {
		text = strings.TrimSpace(text)
		if reFigureNumbered.MatchString(text) || reGraphicGenre.MatchString(text) {
			return true
		}
	}

	return false
}

var (
	reFigureNumbered = regexp.MustCompile(`(?i)^fig(ure|\.)?\s?\d`)

	// Start-anchored with at most an article and one qualifying word before
	// the genre noun ("Bar chart of…", "A Sankey diagram showing…"), so prose
	// that merely mentions one ("Man holding a chart") never matches. Bare
	// "plot" is excluded — "Plot of land…" captions photographs.
	reGraphicGenre = regexp.MustCompile(`(?i)^(an?\s|the\s)?(\S+\s)?(charts?|graphs?|diagrams?|histograms?|flow\s?charts?|heat\s?maps?|(scatter|line|box|bar)\s?plots?)\b`)
)
