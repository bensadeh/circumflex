package comment

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Parse converts one comment's HN HTML into blocks ready for RenderBlocks.
// The whole content pipeline runs here — DOM walk, quote detection, verbatim
// tokens, typography, prose tokens, whitespace — so a comment is parsed once
// no matter how many times the view re-renders it at new widths.
//
// HN's comment grammar is closed: only <p>, <i>, <a href>, <pre><code> and
// entities reach us (user-typed markup arrives entity-escaped).
func Parse(commentHTML string) []Block {
	if Removed(commentHTML) {
		return []Block{{kind: blockRemoved, text: commentHTML}}
	}

	blocks := parseBlocks(commentHTML)

	// Verbatim content — inline code and URLs — is carved out first so
	// typography cannot rewrite it. Quote markers are stripped after
	// typography: a smiley needs its preceding whitespace, so "> :)"
	// converts while ">:)" stays punctuation.
	for i := range blocks {
		detectQuote(&blocks[i])
		tokenizeVerbatim(&blocks[i])
		applyTypography(&blocks[i])
		stripQuoteMarker(&blocks[i])
		trimParagraphLead(&blocks[i])
	}

	tokenizeProse(blocks)
	normalizeWhitespace(blocks)

	return blocks
}

func parseBlocks(src string) []Block {
	ctx := &html.Node{Type: html.ElementNode, DataAtom: atom.Body, Data: "body"}

	nodes, err := html.ParseFragment(strings.NewReader(src), ctx)
	if err != nil {
		return []Block{{kind: blockParagraph, spans: []span{{text: src}}}}
	}

	p := &bodyParser{}
	for _, n := range nodes {
		p.walk(n, spanPlain)
	}

	p.flush()

	return p.blocks
}

// bodyParser walks the comment fragment: <p> separates paragraphs, <pre> is
// a code block of its own, <i> and <a> contribute spans. Boundaries with
// nothing accumulated produce nothing — an empty paragraph has no rendering.
type bodyParser struct {
	blocks []Block
	spans  []span
}

func (p *bodyParser) walk(n *html.Node, format spanFormat) {
	switch n.Type {
	case html.TextNode:
		p.appendText(n.Data, format)

		return

	case html.ElementNode:

	default:
		return
	}

	switch nodeAtom(n) {
	case atom.P:
		p.flush()
		p.walkChildren(n, spanPlain)

	case atom.Pre:
		p.flush()
		p.blocks = append(p.blocks, Block{kind: blockCode, text: textContent(n)})

	case atom.I:
		p.walkChildren(n, spanItalic)

	case atom.A:
		p.appendAnchor(n, format)

	default:
		p.walkChildren(n, format)
	}
}

func (p *bodyParser) walkChildren(n *html.Node, format spanFormat) {
	for c := range n.ChildNodes() {
		p.walk(c, format)
	}
}

func (p *bodyParser) flush() {
	if len(p.spans) == 0 {
		return
	}

	p.blocks = append(p.blocks, Block{kind: blockParagraph, spans: p.spans})
	p.spans = nil
}

func (p *bodyParser) appendText(text string, format spanFormat) {
	if text == "" {
		return
	}

	// Merge adjacent same-format runs so later passes see contiguous text.
	if n := len(p.spans); n > 0 && p.spans[n-1].format == format {
		p.spans[n-1].text += text

		return
	}

	p.spans = append(p.spans, span{text: text, format: format})
}

// appendAnchor emits a link for the anchor's href. HN truncates long URLs in
// the display text, so the href is authoritative for both the target and the
// display. An anchor without a URL href keeps its display text.
func (p *bodyParser) appendAnchor(n *html.Node, format spanFormat) {
	href := attrVal(n, "href")

	if !hasURLScheme(href) {
		p.appendText(textContent(n), format)

		return
	}

	p.spans = append(p.spans, span{text: href, format: spanLink, href: href})
}

func hasURLScheme(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// detectQuote turns a paragraph starting with a > marker into a quote block
// and dissolves its italics: a quote renders faint italic throughout, so an
// <i> run inside one carries no distinction, and folding it up front lets
// every later pass see the contiguous text.
func detectQuote(b *Block) {
	if b.kind != blockParagraph || len(b.spans) == 0 {
		return
	}

	first := b.spans[0]
	if first.format != spanPlain && first.format != spanItalic {
		return
	}

	if !strings.HasPrefix(first.text, ">") && !strings.HasPrefix(first.text, " >") {
		return
	}

	// A > run leading into = or < is a comparison operator (">= 3 versions",
	// ">>="), not the quote convention — stripping its marker would change
	// meaning, so the paragraph stays verbatim.
	rest := strings.TrimLeft(first.text, "> ")
	if strings.HasPrefix(rest, "=") || strings.HasPrefix(rest, "<") {
		return
	}

	b.kind = blockQuote
	foldItalics(b)
}

// foldItalics rewrites italic spans as plain and merges adjacent runs.
func foldItalics(b *Block) {
	folded := b.spans[:0]

	for _, s := range b.spans {
		if s.format == spanItalic {
			s.format = spanPlain
		}

		if n := len(folded); n > 0 && s.format == spanPlain && folded[n-1].format == spanPlain {
			folded[n-1].text += s.text

			continue
		}

		folded = append(folded, s)
	}

	b.spans = folded
}

// stripQuoteMarker removes the leading run of > markers (and the spaces
// around them) from a quote. Nesting depth is not rendered, so ">> x" and
// "> > x" both reduce to their text.
func stripQuoteMarker(b *Block) {
	if b.kind != blockQuote || len(b.spans) == 0 {
		return
	}

	b.spans[0].text = strings.TrimLeft(b.spans[0].text, "> ")
	if b.spans[0].text == "" {
		b.spans = b.spans[1:]
	}
}

// trimParagraphLead drops leading spaces of a paragraph.
func trimParagraphLead(b *Block) {
	if b.kind != blockParagraph || len(b.spans) == 0 || b.spans[0].format != spanPlain {
		return
	}

	b.spans[0].text = strings.TrimLeft(b.spans[0].text, " ")
	if b.spans[0].text == "" {
		b.spans = b.spans[1:]
	}
}

func nodeAtom(n *html.Node) atom.Atom {
	if n.DataAtom != 0 {
		return n.DataAtom
	}

	return atom.Lookup([]byte(n.Data))
}

func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}

	return ""
}

func textContent(n *html.Node) string {
	var sb strings.Builder

	var visit func(*html.Node)

	visit = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}

		for c := range n.ChildNodes() {
			visit(c)
		}
	}
	visit(n)

	return sb.String()
}
