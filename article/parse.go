package article

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/bensadeh/circumflex/ansi"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func parseBlocks(root *html.Node) []block {
	p := &domParser{}
	p.walkChildren(root)
	p.flushInline()

	blocks := dedupeBlocks(p.blocks)
	normalizeHeadings(blocks)

	return blocks
}

// Responsive image markup often leaves the same image or credit line twice
// in readability's output.
func dedupeBlocks(blocks []block) []block {
	var out []block

	for _, b := range blocks {
		if len(out) > 0 {
			prev := out[len(out)-1]
			if prev.kind == b.kind && prev.level == b.level && prev.plainText() == b.plainText() {
				continue
			}
		}

		out = append(out, b)
	}

	return out
}

type domParser struct {
	blocks []block
	inline []span  // pending inline content, flushed as an implicit paragraph
	images []block // images seen in inline flow, emitted after their paragraph
}

// Readability synthesizes elements (e.g. div-to-p conversion) with only the
// tag name set, so DataAtom alone misidentifies them as unknown elements.
func nodeAtom(n *html.Node) atom.Atom {
	if n.DataAtom != 0 {
		return n.DataAtom
	}

	return atom.Lookup([]byte(n.Data))
}

func (p *domParser) walkChildren(n *html.Node) {
	for c := range n.ChildNodes() {
		p.walk(c)
	}
}

func (p *domParser) walk(n *html.Node) {
	switch n.Type {
	case html.TextNode:
		p.inline = append(p.inline, inlineSpans(n, formatPlain, &p.images)...)

		return

	case html.ElementNode:

	default:
		return
	}

	switch nodeAtom(n) {
	case atom.Script, atom.Style, atom.Noscript, atom.Template, atom.Iframe, atom.Head,
		atom.Meta, atom.Link, atom.Title, atom.Form, atom.Button, atom.Input, atom.Select,
		atom.Textarea, atom.Nav, atom.Svg:
		return

	case atom.P:
		p.flushInline()
		p.emitParagraph(collectInline(n, formatPlain, &p.images))
		p.emitImages()

	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		p.flushInline()

		level := int(n.Data[1] - '0')
		text := strings.TrimSpace(spanText(collectInline(n, formatPlain, nil)))

		if text != "" {
			p.blocks = append(p.blocks, block{kind: blockHeading, level: level, text: text})
		}

	case atom.Ul, atom.Ol:
		p.flushInline()

		if items := parseListItems(n, 0, &p.images); len(items) > 0 {
			p.blocks = append(p.blocks, block{kind: blockList, items: items})
		}

		p.emitImages()

	case atom.Pre:
		p.flushInline()

		if text := strings.Trim(preText(n), "\n"); strings.TrimSpace(text) != "" {
			p.blocks = append(p.blocks, block{kind: blockCode, text: text})
		}

	case atom.Blockquote:
		p.flushInline()

		if spans := parseQuote(n); len(spans) > 0 {
			p.blocks = append(p.blocks, block{kind: blockQuote, spans: spans})
		}

	case atom.Table:
		p.flushInline()
		p.parseTable(n)

	case atom.Img:
		p.appendImage(n)

	case atom.Figure:
		p.flushInline()
		p.parseFigure(n)

	case atom.Hr:
		p.flushInline()
		p.blocks = append(p.blocks, block{kind: blockDivider})

	case atom.Br:
		p.inline = append(p.inline, span{text: "\n"})

	case atom.Div, atom.Section, atom.Article, atom.Main, atom.Aside, atom.Header,
		atom.Footer, atom.Details, atom.Summary, atom.Body, atom.Html, atom.Center,
		atom.Dl, atom.Dt, atom.Dd, atom.Li, atom.Figcaption, atom.Fieldset:
		p.flushInline()
		p.walkChildren(n)
		p.flushInline()

	default:
		// Custom elements (e.g. GitHub's table wrappers) land here: treat
		// them as containers when they hold block content, else as inline.
		if hasBlockDescendant(n) {
			p.flushInline()
			p.walkChildren(n)
			p.flushInline()
		} else {
			p.inline = append(p.inline, collectInline(n, formatPlain, &p.images)...)
		}
	}
}

func hasBlockDescendant(n *html.Node) bool {
	for c := range n.Descendants() {
		if c.Type != html.ElementNode {
			continue
		}

		switch nodeAtom(c) {
		case atom.P, atom.Div, atom.Ul, atom.Ol, atom.Table, atom.Pre, atom.Blockquote,
			atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.Figure, atom.Hr,
			atom.Section, atom.Article:
			return true
		}
	}

	return false
}

func (p *domParser) flushInline() {
	spans := p.inline
	p.inline = nil
	p.emitParagraph(spans)
	p.emitImages()
}

func (p *domParser) emitParagraph(spans []span) {
	if spans = normalizeSpans(spans); len(spans) > 0 {
		p.blocks = append(p.blocks, block{kind: blockParagraph, spans: spans})
	}
}

func (p *domParser) emitImages() {
	p.blocks = append(p.blocks, p.images...)
	p.images = nil
}

func (p *domParser) appendImage(n *html.Node) {
	p.images = append(p.images, imageBlock(altText(n)))
}

func imageBlock(caption string) block {
	b := block{kind: blockImage}
	if caption != "" {
		b.spans = []span{{text: caption}}
	}

	return b
}

func (p *domParser) parseFigure(n *html.Node) {
	var img, figcaption *html.Node

	for c := range n.Descendants() {
		if c.Type != html.ElementNode {
			continue
		}

		switch nodeAtom(c) {
		case atom.Img:
			if img == nil {
				img = c
			}

		case atom.Figcaption:
			if figcaption == nil {
				figcaption = c
			}
		}
	}

	if img == nil && figcaption == nil {
		p.walkChildren(n)
		p.flushInline()

		return
	}

	caption := ""
	if figcaption != nil {
		caption = strings.TrimSpace(spanText(normalizeSpans(collectInline(figcaption, formatPlain, nil))))
	}

	if caption == "" && img != nil {
		caption = altText(img)
	}

	p.blocks = append(p.blocks, imageBlock(caption))
}

func (p *domParser) parseTable(n *html.Node) {
	var rows [][]string

	var visitRows func(*html.Node)

	visitRows = func(group *html.Node) {
		for c := range group.ChildNodes() {
			if c.Type != html.ElementNode {
				continue
			}

			switch nodeAtom(c) {
			case atom.Thead, atom.Tbody, atom.Tfoot:
				visitRows(c)

			case atom.Tr:
				if row := tableRow(c); len(row) > 0 {
					rows = append(rows, row)
				}

			case atom.Caption:
				p.emitParagraph(collectInline(c, formatPlain, nil))
			}
		}
	}

	visitRows(n)

	if len(rows) > 0 {
		p.blocks = append(p.blocks, block{kind: blockTable, rows: rows})
	}
}

func tableRow(tr *html.Node) []string {
	var row []string

	empty := true

	for c := range tr.ChildNodes() {
		if c.Type != html.ElementNode || (nodeAtom(c) != atom.Td && nodeAtom(c) != atom.Th) {
			continue
		}

		// Rows render as single lines, so hard breaks inside a cell flatten
		// back to spaces.
		cell := strings.Join(strings.Fields(spanText(normalizeSpans(collectInline(c, formatPlain, nil)))), " ")
		if cell != "" {
			empty = false
		}

		row = append(row, cell)
	}

	if empty {
		return nil
	}

	return row
}

func parseListItems(list *html.Node, depth int, images *[]block) []listItem {
	ordered := nodeAtom(list) == atom.Ol
	number := startNumber(list) - 1

	var items []listItem

	for li := range list.ChildNodes() {
		if li.Type != html.ElementNode || nodeAtom(li) != atom.Li {
			continue
		}

		if ordered {
			number++
		}

		var spans []span

		var nested []listItem

		for c := range li.ChildNodes() {
			if c.Type == html.ElementNode && (nodeAtom(c) == atom.Ul || nodeAtom(c) == atom.Ol) {
				nested = append(nested, parseListItems(c, depth+1, images)...)

				continue
			}

			spans = append(spans, span{text: " "})
			spans = append(spans, inlineSpans(c, formatPlain, images)...)
		}

		if spans = normalizeSpans(spans); len(spans) > 0 {
			item := listItem{depth: depth, spans: spans}
			if ordered {
				item.number = number
			}

			items = append(items, item)
		}

		items = append(items, nested...)
	}

	return items
}

func startNumber(list *html.Node) int {
	if start, err := strconv.Atoi(attr(list, "start")); err == nil && start > 0 {
		return start
	}

	return 1
}

func parseQuote(n *html.Node) []span {
	sub := &domParser{}
	sub.walkChildren(n)
	sub.flushInline()

	var spans []span

	for _, b := range sub.blocks {
		var line []span

		switch b.kind {
		case blockParagraph, blockQuote:
			line = b.spans

		case blockHeading, blockCode, blockVerbatim:
			line = []span{{text: b.text}}

		case blockList:
			for _, item := range b.items {
				line = append(line, span{text: "- "})
				line = append(line, item.spans...)
				line = append(line, span{text: "\n"})
			}

			if len(line) > 0 {
				line = line[:len(line)-1]
			}

		case blockTable, blockImage, blockDivider:
			continue

		default:
			continue
		}

		if len(spans) > 0 {
			spans = append(spans, span{text: "\n\n"})
		}

		spans = append(spans, line...)
	}

	return spans
}

func preText(n *html.Node) string {
	var sb strings.Builder

	var visit func(*html.Node)

	visit = func(c *html.Node) {
		switch {
		case c.Type == html.TextNode:
			sb.WriteString(ansi.Strip(c.Data))

		case c.Type == html.ElementNode && nodeAtom(c) == atom.Br:
			sb.WriteByte('\n')

		case c.Type == html.ElementNode:
			for gc := range c.ChildNodes() {
				visit(gc)
			}
		}
	}

	for c := range n.ChildNodes() {
		visit(c)
	}

	// The terminal expands a tab to the next 8-column stop, but the wrapper
	// counts it as one cell; expand here so widths agree, as the text/plain
	// path does.
	return strings.ReplaceAll(sb.String(), "\t", strings.Repeat(" ", 8))
}

// Bold has no case below on purpose: it unwraps to plain text.
// A nil images sink discards images instead of collecting them.
func collectInline(n *html.Node, format inlineFormat, images *[]block) []span {
	var spans []span

	for c := range n.ChildNodes() {
		spans = append(spans, inlineSpans(c, format, images)...)
	}

	return spans
}

func inlineSpans(n *html.Node, format inlineFormat, images *[]block) []span {
	switch n.Type {
	case html.TextNode:
		text := collapseWhitespace(ansi.Strip(n.Data))
		if format != formatCode {
			text = strings.ReplaceAll(text, "...", "…")
		}

		return []span{{text: text, format: format}}

	case html.ElementNode:

	default:
		return nil
	}

	switch nodeAtom(n) {
	case atom.Script, atom.Style, atom.Noscript, atom.Template, atom.Svg:
		return nil

	case atom.Br:
		return []span{{text: "\n", format: format}}

	case atom.Img:
		if images != nil {
			*images = append(*images, imageBlock(altText(n)))
		}

		return nil

	case atom.Em, atom.I, atom.Var, atom.Dfn:
		if format == formatPlain {
			format = formatItalic
		}

		return collectInline(n, format, images)

	case atom.Del, atom.S, atom.Strike:
		if format == formatPlain {
			format = formatStrike
		}

		return collectInline(n, format, images)

	case atom.Code, atom.Kbd, atom.Samp:
		return collectInline(n, formatCode, images)

	case atom.A:
		return linkSpans(n, format, images)

	case atom.Sup:
		return scriptSpans(n, format, images, superscriptRunes)

	case atom.Sub:
		return scriptSpans(n, format, images, subscriptRunes)

	case atom.P, atom.Div, atom.Li:
		spans := []span{{text: " ", format: format}}
		spans = append(spans, collectInline(n, format, images)...)

		return append(spans, span{text: " ", format: format})

	default:
		return collectInline(n, format, images)
	}
}

// linkSpans unwraps an anchor to its text but keeps the target as an OSC 8
// hyperlink, so links stay clickable without changing how they look.
func linkSpans(n *html.Node, format inlineFormat, images *[]block) []span {
	spans := collectInline(n, format, images)

	href := attr(n, "href")
	if !isLinkableHref(href) {
		return spans
	}

	for i := range spans {
		spans[i].href = href
	}

	return spans
}

// isLinkableHref accepts only absolute http(s) URLs free of control characters,
// so an attacker-supplied href cannot terminate the OSC 8 hyperlink sequence
// early and inject terminal escapes.
func isLinkableHref(href string) bool {
	if !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "https://") {
		return false
	}

	return !strings.ContainsFunc(href, unicode.IsControl)
}

var superscriptRunes = map[rune]rune{
	'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴',
	'5': '⁵', '6': '⁶', '7': '⁷', '8': '⁸', '9': '⁹',
	'+': '⁺', '-': '⁻', '=': '⁼', '(': '⁽', ')': '⁾', 'n': 'ⁿ', 'i': 'ⁱ',
}

var subscriptRunes = map[rune]rune{
	'0': '₀', '1': '₁', '2': '₂', '3': '₃', '4': '₄',
	'5': '₅', '6': '₆', '7': '₇', '8': '₈', '9': '₉',
	'+': '₊', '-': '₋', '=': '₌', '(': '₍', ')': '₎',
}

// scriptSpans converts sup/sub content to Unicode equivalents when every rune
// has one, and otherwise leaves the content as regular inline text.
func scriptSpans(n *html.Node, format inlineFormat, images *[]block, mapping map[rune]rune) []span {
	spans := collectInline(n, format, images)

	text := spanText(spans)
	if text == "" {
		return spans
	}

	var sb strings.Builder

	for _, r := range text {
		mapped, ok := mapping[r]
		if !ok {
			return spans
		}

		sb.WriteRune(mapped)
	}

	return []span{{text: sb.String(), format: format}}
}

// Whitespace between spans collapses to a single separator: a hard break
// (from <br>) wins over a space, and runs of breaks cap at one blank line.
// Edge whitespace is dropped, so paragraphs never start or end blank.
func normalizeSpans(spans []span) []span {
	var out []span

	emit := func(s span) {
		if len(out) > 0 && out[len(out)-1].format == s.format && out[len(out)-1].href == s.href {
			out[len(out)-1].text += s.text
		} else {
			out = append(out, s)
		}
	}

	newlines := 0
	space := false

	for _, s := range spans {
		text := strings.TrimLeft(s.text, " \n")
		lead := s.text[:len(s.text)-len(text)]

		newlines += strings.Count(lead, "\n")
		space = space || strings.Contains(lead, " ")

		if text == "" {
			continue
		}

		trimmed := strings.TrimRight(text, " \n")
		trail := text[len(trimmed):]

		if len(out) > 0 {
			switch {
			case newlines > 0:
				emit(span{text: strings.Repeat("\n", min(newlines, 2))})
			case space:
				emit(span{text: " "})
			}
		}

		newlines = strings.Count(trail, "\n")
		space = strings.Contains(trail, " ")

		emit(span{text: trimmed, format: s.format, href: s.href})
	}

	return out
}

var invisibleChars = strings.NewReplacer(
	"\u200b", "", // zero-width space
	"\ufeff", "", // byte order mark
	"\u00ad", "", // soft hyphen
)

// Edge whitespace collapses to a single space rather than nothing so that
// spacing between adjacent nodes survives.
func collapseWhitespace(s string) string {
	if s == "" {
		return ""
	}

	s = invisibleChars.Replace(s)

	collapsed := strings.Join(strings.Fields(s), " ")
	if collapsed == "" {
		return " "
	}

	if unicode.IsSpace(firstRune(s)) {
		collapsed = " " + collapsed
	}

	if unicode.IsSpace(lastRune(s)) {
		collapsed += " "
	}

	return collapsed
}

func firstRune(s string) rune {
	for _, r := range s {
		return r
	}

	return 0
}

func lastRune(s string) rune {
	var last rune
	for _, r := range s {
		last = r
	}

	return last
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}

	return ""
}

// altText reads an img alt attribute. Attribute values are stripped of control
// sequences like text nodes are, since the source is equally untrusted.
func altText(n *html.Node) string {
	return strings.TrimSpace(collapseWhitespace(ansi.Strip(attr(n, "alt"))))
}

// Heading levels are remapped to a contiguous 1..n range so that articles
// starting at h3 still render as top-level sections.
func normalizeHeadings(blocks []block) {
	var seen [7]bool

	for i := range blocks {
		if blocks[i].kind == blockHeading {
			seen[blocks[i].level] = true
		}
	}

	var mapping [7]int

	next := 1

	for level := 1; level <= 6; level++ {
		if seen[level] {
			mapping[level] = next
			next++
		}
	}

	for i := range blocks {
		if blocks[i].kind == blockHeading {
			blocks[i].level = mapping[blocks[i].level]
		}
	}
}
