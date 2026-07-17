package list

import (
	"sort"
	"strings"

	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/view/pane"

	xansi "github.com/charmbracelet/x/ansi"
)

// highlightQuery repaints the query's matches in the already-styled title,
// with exactly the in-page search semantics: each query word matches as a
// smart-case substring, so "test" lights up inside "testing" here just as it
// does in a comment section. The selected row paints in the current-match
// tier, the rest in the all-matches tier. Matching against the visible text
// keeps upstream token transforms (glyph swaps, dropped parens) from
// desyncing the spans.
func highlightQuery(title, query string, current bool) string {
	spans := querySpans(xansi.Strip(title), query, current)
	if len(spans) == 0 {
		return title
	}

	return style.OverlaySearchSpans(title, spans)
}

func querySpans(plain, query string, current bool) []style.SearchSpan {
	var spans []style.SearchSpan

	for word := range strings.FieldsSeq(query) {
		for _, m := range pane.FindMatches([]string{plain}, word) {
			spans = append(spans, style.SearchSpan{
				StartCell: m.StartCell,
				EndCell:   m.EndCell,
				Current:   current,
			})
		}
	}

	sort.Slice(spans, func(i, j int) bool { return spans[i].StartCell < spans[j].StartCell })

	return dropOverlappingSpans(spans)
}

// dropOverlappingSpans keeps the sorted spans' first claim on any cell;
// OverlaySearchSpans requires non-overlapping input. Overlaps only arise
// when one query word contains another, so dropping the latecomer is fine.
func dropOverlappingSpans(spans []style.SearchSpan) []style.SearchSpan {
	out := spans[:0]
	end := -1

	for _, sp := range spans {
		if sp.StartCell < end {
			continue
		}

		out = append(out, sp)
		end = sp.EndCell
	}

	return out
}
