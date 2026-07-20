package pane

import (
	"image/color"
	"strings"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"

	xansi "github.com/charmbracelet/x/ansi"
)

// Link is one followable URL in a rendered document: the OSC 8 target and
// the cell spans its anchor text occupies — one span per rendered line for
// links the wrap split.
type Link struct {
	URL      string
	Spans    []Match
	Viewable bool
}

// LinkViewable reports whether following the link can open a view in place:
// a page reader mode could render, or a Hacker News discussion the comment
// section opens. Links that can't (PDFs, media, archives, blocked domains)
// are selectable but inert, marked by the muted selection bar and a dimmed
// footer URL.
func LinkViewable(rawURL string) bool {
	if _, ok := hn.ParseItemURL(rawURL); ok {
		return true
	}

	return article.ValidateURL(rawURL) == nil
}

const linkOpenMarker = "\x1b]8;;"

// ExtractLinks locates every OSC 8 hyperlink in the rendered lines from
// fromLine on. The wrapper closes and reopens a hyperlink at each line
// break, so spans of the same URL on adjacent lines are one link, while the
// same URL opening twice on one line — or in separate paragraphs, a blank
// line apart — is two.
func ExtractLinks(lines []string, fromLine int) []Link {
	var links []Link

	for lineIdx := max(0, fromLine); lineIdx < len(lines); lineIdx++ {
		scanLine(lines[lineIdx], lineIdx, &links)
	}

	return links
}

// scanLine walks one line marker to marker, accumulating cell offsets
// segment by segment like FindMatches — a prefix width per marker would be
// quadratic on link-dense lines.
func scanLine(line string, lineIdx int, links *[]Link) {
	pos, cell := 0, 0
	url, spanStart := "", 0

	flush := func(endCell int) {
		if url == "" || endCell <= spanStart {
			return
		}

		m := Match{Line: lineIdx, StartCell: spanStart, EndCell: endCell}

		if n := len(*links); n > 0 {
			last := &(*links)[n-1]
			lastSpan := last.Spans[len(last.Spans)-1]

			if last.URL == url && lastSpan.Line == lineIdx-1 {
				last.Spans = append(last.Spans, m)

				return
			}
		}

		*links = append(*links, Link{URL: url, Spans: []Match{m}, Viewable: LinkViewable(url)})
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

// LinkSelectorLabel is the footer label while a URL selector is up: the
// selector icon and a faint mode label, while the URL itself rides the
// separator above. A link the view won't open breaks the arrow, matching its
// dimmed URL; an empty selection keeps the plain arrow — nothing is inert,
// there is just nothing selected yet.
func LinkSelectorLabel(viewable, nerdFonts bool) string {
	icon, sep := style.Faint("→"), " "
	if !viewable {
		icon = style.Faint("↛")
	}

	if nerdFonts {
		// Nerd font glyphs render wider than one cell, so they get extra room.
		icon, sep = nerdfonts.LinkSelector, "  "
		if !viewable {
			icon = nerdfonts.LinkSelectorOff
		}
	}

	return "  " + icon + sep + style.Faint("URL Selection Mode")
}

// LinkURLRow is the footer separator while a URL selector is up: the
// selected link's URL written into the rule, right-aligned against the
// content column's edge, leaving the footer line to the mode label. The URL
// renders faint under the rule's full-strength underline; whether the link
// opens shows in the footer arrow and the selection bar, not here.
func LinkURLRow(paneWidth, leftMargin, contentWidth int, url string, termFG, termBG color.Color) string {
	// The scheme is stripped from the display like the meta block's URL row —
	// the selector is visibly showing a link already.
	display := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	display = xansi.Truncate(display, contentWidth, "…")

	return FooterSeparatorWithLabel(paneWidth, leftMargin+contentWidth, display, termFG, termBG)
}

// LinkOnScreen reports whether any of the link's spans lies in the viewport
// window starting at top, height lines tall.
func LinkOnScreen(l Link, top, height int) bool {
	for _, s := range l.Spans {
		if s.Line >= top && s.Line < top+height {
			return true
		}
	}

	return false
}

// FirstLinkOnScreen is the topmost link in the viewport window, or -1 with
// none in view.
func FirstLinkOnScreen(links []Link, top, height int) int {
	for i := range links {
		if LinkOnScreen(links[i], top, height) {
			return i
		}
	}

	return -1
}

// MoveLink selects the next link wherever it sits, wrapping like a search
// jump. From an empty selection (-1) it leaves relative to the viewport:
// forward to the first link past its top, backward to the last one above it.
func MoveLink(links []Link, current, direction, top int) int {
	n := len(links)

	switch {
	case current >= 0:
		return ((current+direction)%n + n) % n

	case direction > 0:
		for i, l := range links {
			if l.Spans[0].Line >= top {
				return i
			}
		}

		return 0

	default:
		for i := n - 1; i >= 0; i-- {
			if links[i].Spans[0].Line < top {
				return i
			}
		}

		return n - 1
	}
}
