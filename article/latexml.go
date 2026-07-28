package article

import (
	nurl "net/url"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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

// normalizeLatexmlTables drops the ltx_guessed_headers marker LaTeXML puts on
// tables whose header rows it inferred: readability's unlikely-candidate
// regex reads the "header" substring as page chrome and deletes the whole
// table before scoring ever runs. The marker is converter metadata with no
// reader meaning. Only this token is scrubbed — ltx_page_header and friends
// really are chrome, and stay deletable.
func normalizeLatexmlTables(doc *html.Node) {
	for n := range doc.Descendants() {
		if n.Type != html.ElementNode || !hasClass(n, "ltx_guessed_headers") {
			continue
		}

		for i, a := range n.Attr {
			if a.Key == "class" {
				classes := slices.DeleteFunc(strings.Fields(a.Val), func(c string) bool {
					return c == "ltx_guessed_headers"
				})
				n.Attr[i].Val = strings.Join(classes, " ")
			}
		}
	}
}

// dropLatexmlRawPictures empties ltx_picture elements holding a picture
// LaTeXML could not convert: the raw LaTeX source — \begin{overpic} and its
// \put commands — is left as the span's text, and the graphic it references
// is never copied into the paper's asset tree, so there is no image to
// recover. Emptied, the enclosing figure collapses to its caption and takes
// the Figure designation instead of rendering pages of markup. A converted
// picture keeps its img (or svg) child and carries no raw source, and is
// left alone.
func dropLatexmlRawPictures(doc *html.Node) {
	var raw []*html.Node

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && hasClass(n, "ltx_picture") && isRawLatexPicture(n) {
			raw = append(raw, n)
		}
	}

	for _, n := range raw {
		for n.FirstChild != nil {
			n.RemoveChild(n.FirstChild)
		}
	}
}

func isRawLatexPicture(n *html.Node) bool {
	for c := range n.Descendants() {
		if c.Type == html.ElementNode && (nodeAtom(c) == atom.Img || nodeAtom(c) == atom.Svg) {
			return false
		}
	}

	return strings.HasPrefix(strings.TrimSpace(nodeText(n)), `\begin{`)
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
