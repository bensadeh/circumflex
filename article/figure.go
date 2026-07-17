package article

import (
	"regexp"
	"strings"
)

// knownFigure reports whether an image is known to depict a drawn graphic — a
// chart, plot, or diagram — where half-block art degrades to an illegible
// smear and the description carries the actual content. Only declared
// knowledge counts, so photographs are never demoted: a vector source (svg is
// drawn by definition, never photographed), a print-style numbered caption
// ("Figure 3: …"), or a description led by the graphic's genre, the alt-text
// convention for complex images ("Bar chart of …"). A vector source with no
// description at all stays an image: art may be a smear, but a bare label
// says even less.
func knownFigure(src string, texts ...string) bool {
	described := false

	for _, text := range texts {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		described = true

		if reFigureNumbered.MatchString(text) || reGraphicGenre.MatchString(text) {
			return true
		}
	}

	return described && isVectorURL(src)
}

var (
	reFigureNumbered = regexp.MustCompile(`(?i)^fig(ure|\.)?\s?\d`)

	// Start-anchored with at most an article and one qualifying word before
	// the genre noun ("Bar chart of…", "A Sankey diagram showing…"), so prose
	// that merely mentions one ("Man holding a chart") never matches. Bare
	// "plot" is excluded — "Plot of land…" captions photographs.
	reGraphicGenre = regexp.MustCompile(`(?i)^(an?\s|the\s)?(\S+\s)?(charts?|graphs?|diagrams?|histograms?|flow\s?charts?|heat\s?maps?|(scatter|line|box|bar)\s?plots?)\b`)
)

func isVectorURL(src string) bool {
	path := src
	if i := strings.IndexAny(path, "?#"); i >= 0 {
		path = path[:i]
	}

	return strings.HasSuffix(strings.ToLower(path), ".svg")
}
