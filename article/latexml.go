package article

import (
	nurl "net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// LaTeXML is the LaTeX-to-HTML converter behind arXiv's paper renderings;
// its markup is told apart by ltx_-prefixed class names. Everything arXiv-
// and LaTeXML-specific lives here: the full-text mirror lookup and the
// footnote folding. (The abstract page's trim rules stay data in sites.go's
// table with every other site's.)

var arxivPagePath = regexp.MustCompile(`^/(?:abs|pdf)/(.+?)(?:\.pdf)?/?$`)

// fullTextURL returns a known full-text rendering of the page at u, or "".
// arXiv abstract and PDF links map onto /html/<id>, an HTML version of the
// paper generated from its LaTeX source: reading that beats the abstract-only
// /abs page, and /pdf reader mode cannot parse at all. Papers without a
// conversion return 404 there, and the fetch falls back to the original URL.
func fullTextURL(u *nurl.URL) string {
	host := strings.TrimPrefix(u.Hostname(), "www.")
	if host != "arxiv.org" && host != "export.arxiv.org" {
		return ""
	}

	match := arxivPagePath.FindStringSubmatch(u.EscapedPath())
	if match == nil {
		return ""
	}

	return "https://arxiv.org/html/" + match[1]
}

// latexmlPreservedClasses names the footnote chrome readability must not
// strip, so the parser can fold the popup markup into a readable form.
var latexmlPreservedClasses = []string{
	"ltx_note", "ltx_note_mark", "ltx_note_type", "ltx_note_content", "ltx_tag_note",
}

func isLatexmlNote(n *html.Node) bool {
	return hasClass(n, "ltx_note")
}

// noteSpans renders a LaTeXML footnote, whose markup carries popup chrome: the
// superscript mark appears twice (outside and inside the note body), joined by
// a "footnotemark: " label and a tag number. The note text reads best inline
// as a parenthetical; a bare mark with no text (\footnotemark) keeps only its
// superscript number.
func noteSpans(n *html.Node, format inlineFormat, images *[]block) []span {
	content := descendantWithClass(n, "ltx_note_content")
	if content == nil {
		return collectInline(n, format, images)
	}

	var spans []span

	for c := range content.ChildNodes() {
		if c.Type == html.ElementNode &&
			(hasClass(c, "ltx_note_mark") || hasClass(c, "ltx_note_type") || hasClass(c, "ltx_tag_note")) {
			continue
		}

		spans = append(spans, inlineSpans(c, format, images)...)
	}

	if len(normalizeSpans(spans)) == 0 {
		if mark := descendantWithClass(n, "ltx_note_mark"); mark != nil {
			return scriptSpans(mark, format, nil, superscriptRunes)
		}

		return nil
	}

	out := []span{{text: " (", format: format}}
	out = append(out, spans...)

	return append(out, span{text: ")", format: format})
}
