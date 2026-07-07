package article

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/bensadeh/circumflex/ansi"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// parseBlocks converts readability's cleaned DOM into the block representation.
func parseBlocks(root *html.Node) []block {
	p := &domParser{}
	p.walkChildren(root)
	p.flushInline()
	normalizeHeadings(p.blocks)

	return p.blocks
}

type domParser struct {
	blocks []block
	inline []span  // pending inline content, flushed as an implicit paragraph
	images []block // images seen in inline flow, emitted after their paragraph
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

	switch n.DataAtom {
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
		p.inline = append(p.inline, span{text: " "})

	case atom.Div, atom.Section, atom.Article, atom.Main, atom.Aside, atom.Header,
		atom.Footer, atom.Details, atom.Summary, atom.Body, atom.Html, atom.Center,
		atom.Dl, atom.Dt, atom.Dd, atom.Li, atom.Figcaption, atom.Fieldset:
		p.flushInline()
		p.walkChildren(n)
		p.flushInline()

	default:
		p.inline = append(p.inline, collectInline(n, formatPlain, &p.images)...)
	}
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
	p.images = append(p.images, imageBlock(strings.TrimSpace(collapseWhitespace(attr(n, "alt")))))
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

		switch c.DataAtom {
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
		caption = strings.TrimSpace(collapseWhitespace(attr(img, "alt")))
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

			switch c.DataAtom {
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
		if c.Type != html.ElementNode || (c.DataAtom != atom.Td && c.DataAtom != atom.Th) {
			continue
		}

		cell := strings.TrimSpace(spanText(normalizeSpans(collectInline(c, formatPlain, nil))))
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
	ordered := list.DataAtom == atom.Ol
	number := startNumber(list) - 1

	var items []listItem

	for li := range list.ChildNodes() {
		if li.Type != html.ElementNode || li.DataAtom != atom.Li {
			continue
		}

		if ordered {
			number++
		}

		var spans []span

		var nested []listItem

		for c := range li.ChildNodes() {
			if c.Type == html.ElementNode && (c.DataAtom == atom.Ul || c.DataAtom == atom.Ol) {
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

// parseQuote flattens a blockquote's content into spans, one line per inner block.
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

		case blockHeading, blockCode:
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
			spans = append(spans, span{text: "\n"})
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

		case c.Type == html.ElementNode && c.DataAtom == atom.Br:
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

	return sb.String()
}

// collectInline flattens a node's content into styled spans. Anchors and bold
// are unwrapped to plain text; images are diverted to the images sink so they
// can be emitted as their own blocks (nil discards them).
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

	switch n.DataAtom {
	case atom.Script, atom.Style, atom.Noscript, atom.Template, atom.Svg:
		return nil

	case atom.Br:
		return []span{{text: " ", format: format}}

	case atom.Img:
		if images != nil {
			*images = append(*images, imageBlock(strings.TrimSpace(collapseWhitespace(attr(n, "alt")))))
		}

		return nil

	case atom.Em, atom.I, atom.Var, atom.Dfn:
		if format == formatPlain {
			format = formatItalic
		}

		return collectInline(n, format, images)

	case atom.Code, atom.Kbd, atom.Samp:
		return collectInline(n, formatCode, images)

	case atom.P, atom.Div, atom.Li:
		spans := []span{{text: " ", format: format}}
		spans = append(spans, collectInline(n, format, images)...)

		return append(spans, span{text: " ", format: format})

	default:
		return collectInline(n, format, images)
	}
}

// normalizeSpans merges adjacent same-format spans, collapses whitespace
// across span boundaries, and trims the edges of the sequence.
func normalizeSpans(spans []span) []span {
	var out []span

	prevSpace := true

	for _, s := range spans {
		text := s.text
		if prevSpace {
			text = strings.TrimPrefix(text, " ")
		}

		if text == "" {
			continue
		}

		prevSpace = strings.HasSuffix(text, " ")

		if len(out) > 0 && out[len(out)-1].format == s.format {
			out[len(out)-1].text += text
		} else {
			out = append(out, span{text: text, format: s.format})
		}
	}

	for len(out) > 0 {
		last := &out[len(out)-1]

		last.text = strings.TrimRight(last.text, " ")
		if last.text != "" {
			break
		}

		out = out[:len(out)-1]
	}

	return out
}

// collapseWhitespace applies HTML whitespace rules to a text node: runs of
// whitespace become a single space, preserved at the edges so that spacing
// between adjacent nodes survives.
func collapseWhitespace(s string) string {
	if s == "" {
		return ""
	}

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

// normalizeHeadings remaps the heading levels in use to a contiguous 1..n
// range so that articles starting at h3 still render as top-level sections.
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
