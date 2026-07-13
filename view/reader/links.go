package reader

import (
	"strings"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/view/pane"

	xansi "github.com/charmbracelet/x/ansi"
)

// link is one followable URL in the rendered article: the OSC 8 target and
// the cell spans its anchor text occupies — one span per rendered line for
// links the wrap split.
type link struct {
	url      string
	spans    []pane.Match
	viewable bool
}

// linkViewable reports whether reader mode could render the target — links
// it can't (PDFs, media, archives, blocked domains) are selectable but
// inert, marked by the muted selection bar and a dimmed footer URL.
func linkViewable(rawURL string) bool {
	return article.ValidateURL(rawURL) == nil
}

const linkOpenMarker = "\x1b]8;;"

// extractLinks locates every OSC 8 hyperlink in the rendered lines from
// fromLine on — the article body; earlier lines belong to the meta header,
// whose URL row is not a selectable link. The wrapper closes and reopens a
// hyperlink at each line break, so spans of the same URL on adjacent lines
// are one link, while the same URL opening twice on one line — or in
// separate paragraphs, a blank line apart — is two.
func extractLinks(lines []string, fromLine int) []link {
	var links []link

	for lineIdx := max(0, fromLine); lineIdx < len(lines); lineIdx++ {
		scanLine(lines[lineIdx], lineIdx, &links)
	}

	return links
}

// scanLine walks one line marker to marker, accumulating cell offsets
// segment by segment like pane.FindMatches — a prefix width per marker would
// be quadratic on link-dense lines.
func scanLine(line string, lineIdx int, links *[]link) {
	pos, cell := 0, 0
	url, spanStart := "", 0

	flush := func(endCell int) {
		if url == "" || endCell <= spanStart {
			return
		}

		m := pane.Match{Line: lineIdx, StartCell: spanStart, EndCell: endCell}

		if n := len(*links); n > 0 {
			last := &(*links)[n-1]
			lastSpan := last.spans[len(last.spans)-1]

			if last.url == url && lastSpan.Line == lineIdx-1 {
				last.spans = append(last.spans, m)

				return
			}
		}

		*links = append(*links, link{url: url, spans: []pane.Match{m}, viewable: linkViewable(url)})
	}

	for {
		idx := strings.Index(line[pos:], linkOpenMarker)
		if idx < 0 {
			flush(cell + xansi.StringWidth(line[pos:]))

			return
		}

		idx += pos
		cell += xansi.StringWidth(line[pos:idx])
		flush(cell)

		target, next := parseLinkTarget(line, idx+len(linkOpenMarker))
		url, spanStart = target, cell
		pos = next
	}
}

// parseLinkTarget reads an OSC 8 target and returns it with the byte offset
// past the sequence terminator. Both terminators appear in rendered output:
// ST from ansi.Hyperlink, BEL from the sequences the wrapper reopens. An
// empty target is a close.
func parseLinkTarget(line string, from int) (string, int) {
	rest := line[from:]

	if bel := strings.IndexByte(rest, '\a'); bel >= 0 {
		if st := strings.Index(rest[:bel], "\x1b\\"); st >= 0 {
			return rest[:st], from + st + 2
		}

		return rest[:bel], from + bel + 1
	}

	if st := strings.Index(rest, "\x1b\\"); st >= 0 {
		return rest[:st], from + st + 2
	}

	return "", len(line)
}
