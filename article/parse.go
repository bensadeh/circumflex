package article

import (
	"slices"
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

			same := prev.kind == b.kind && prev.level == b.level && prev.plainText() == b.plainText()
			if same && b.kind == blockImage && b.plainText() == "" {
				// Uncaptioned images all share an empty caption, so fall back
				// to the source to avoid collapsing genuinely distinct images.
				same = prev.imageURL == b.imageURL
			}

			if same {
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

	if skippedElement(nodeAtom(n)) {
		return
	}

	// Footnotes flow inline even when their content holds block markup, so
	// the popup chrome never splits the surrounding paragraph.
	if isLatexmlNote(n) {
		p.inline = append(p.inline, noteSpans(n, formatPlain, &p.images)...)

		return
	}

	switch nodeAtom(n) {
	case atom.P:
		p.flushInline()

		// Quirks-mode pages (no doctype) keep a <table> nested inside an
		// open <p> instead of auto-closing it, so a paragraph can hold real
		// block content; contain it rather than flattening it to text.
		if hasBlockDescendant(n) {
			p.walkChildren(n)
			p.flushInline()
		} else {
			p.emitParagraph(collectInline(n, formatPlain, &p.images))
			p.emitImages()
		}

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
		if tex := mathFallbackTeX(n); tex != "" {
			p.inline = append(p.inline, span{text: latexToUnicode(tex)})
		} else {
			p.appendImage(n)
		}

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
		// Inline and custom elements (e.g. GitHub's table wrappers) land
		// here: treat them as containers when they hold block content, else
		// as inline flow so formatting tags keep their styling.
		if hasBlockDescendant(n) {
			p.flushInline()
			p.walkChildren(n)
			p.flushInline()
		} else {
			p.inline = append(p.inline, inlineSpans(n, formatPlain, &p.images)...)
		}
	}
}

// skippedElement reports whether the walker drops this element and its
// subtree entirely; anything checking what the walker would render must
// apply the same filter.
func skippedElement(a atom.Atom) bool {
	switch a {
	case atom.Script, atom.Style, atom.Noscript, atom.Template, atom.Iframe, atom.Head,
		atom.Meta, atom.Link, atom.Title, atom.Form, atom.Button, atom.Input, atom.Select,
		atom.Textarea, atom.Nav, atom.Svg:
		return true
	}

	return false
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
	p.images = append(p.images, imageBlock(altText(n), imageSrc(n), imageDisplayWidth(n)))
}

func imageBlock(caption, src string, dispWidth int) block {
	b := block{kind: blockImage, imageURL: src, dispWidth: dispWidth}
	if caption != "" {
		b.spans = []span{{text: caption}}
	}

	return b
}

// imageDisplayWidth reads the img width attribute, the site's intended display
// size in CSS px. A missing or non-positive value returns 0 so rendering falls
// back to the image's intrinsic resolution.
func imageDisplayWidth(n *html.Node) int {
	if n == nil {
		return 0
	}

	if w, err := strconv.Atoi(strings.TrimSpace(attr(n, "width"))); err == nil && w > 0 {
		return w
	}

	return 0
}

// imageSrc picks the most promising source for an img: a right-sized srcset
// variant when the set advertises widths, then the eager attributes, then the
// largest srcset candidate. Lazy-loaded images often hold a placeholder in
// src and the real URL in a data-* attr.
func imageSrc(n *html.Node) string {
	if v := rightSizedFromSrcset(attr(n, "srcset")); v != "" {
		return v
	}

	for _, key := range []string{"src", "data-src", "data-original", "data-lazy-src"} {
		if v := strings.TrimSpace(attr(n, key)); isFetchableImageURL(v) {
			return v
		}
	}

	return bestFromSrcset(attr(n, "srcset"))
}

type srcsetCandidate struct {
	url        string
	descriptor string
}

const srcsetWhitespace = " \t\n\r\f"

// splitSrcset splits a srcset attribute into url/descriptor candidates.
// Candidates are comma-separated, but the URLs themselves may contain commas
// (Substack's CDN encodes transforms as ",w_848,c_limit,…" path segments), so
// a bare split on "," shreds them. Per the HTML spec a URL runs to the next
// whitespace; a comma ends a candidate only glued to the URL's tail
// (descriptor-less candidate) or after the descriptor.
func splitSrcset(srcset string) []srcsetCandidate {
	var candidates []srcsetCandidate

	rest := srcset
	for {
		rest = strings.TrimLeft(rest, srcsetWhitespace+",")
		if rest == "" {
			return candidates
		}

		url := rest
		rest = ""

		if i := strings.IndexAny(url, srcsetWhitespace); i >= 0 {
			url, rest = url[:i], url[i:]
		}

		if trimmed := strings.TrimRight(url, ","); trimmed != url {
			candidates = append(candidates, srcsetCandidate{url: trimmed})

			continue
		}

		descriptor := rest
		rest = ""

		if i := strings.IndexByte(descriptor, ','); i >= 0 {
			descriptor, rest = descriptor[:i], descriptor[i+1:]
		}

		candidates = append(candidates, srcsetCandidate{url: url, descriptor: strings.TrimSpace(descriptor)})
	}
}

// rightSizedFromSrcset returns the smallest width-annotated candidate that
// still covers maxRetainedPx: anything larger is downloaded only to be thrown
// away by boundImage, and a full-size WordPress original runs ~5x the bytes
// of its 768w variant. Returns "" when no candidate is both usable and large
// enough, leaving the eager-attribute chain to decide.
func rightSizedFromSrcset(srcset string) string {
	var best string

	bestWidth := 0

	for _, candidate := range splitSrcset(srcset) {
		if candidate.descriptor == "" || !isFetchableImageURL(candidate.url) {
			continue
		}

		width, err := strconv.Atoi(strings.TrimSuffix(strings.Fields(candidate.descriptor)[0], "w"))
		if err != nil || width < maxRetainedPx {
			continue
		}

		if bestWidth == 0 || width < bestWidth {
			best, bestWidth = candidate.url, width
		}
	}

	return best
}

// data: URIs, inline SVG, and lazy-load placeholders are skipped so imageSrc
// falls through to the real source.
func isFetchableImageURL(v string) bool {
	return v != "" && !strings.HasPrefix(v, "data:") && !isPlaceholderURL(v)
}

// isPlaceholderURL flags the blank/grey spacer images sites show before the
// real image lazy-loads (e.g. BBC's grey-placeholder.png).
func isPlaceholderURL(v string) bool {
	lower := strings.ToLower(v)
	for _, marker := range placeholderMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}

	return false
}

var placeholderMarkers = []string{
	"placeholder", "spacer.gif", "spacer.png", "blank.gif", "blank.png",
	"transparent.gif", "transparent.png", "1x1.gif", "1x1.png",
}

// bestFromSrcset returns the last usable candidate, which is conventionally the
// highest resolution in a "url descriptor, url descriptor" list.
func bestFromSrcset(srcset string) string {
	for _, candidate := range slices.Backward(splitSrcset(srcset)) {
		if isFetchableImageURL(candidate.url) {
			return candidate.url
		}
	}

	return ""
}

func (p *domParser) parseFigure(n *html.Node) {
	var img, figcaption *html.Node

	for c := range n.Descendants() {
		if c.Type != html.ElementNode {
			continue
		}

		switch nodeAtom(c) {
		case atom.Img:
			// Prefer an img with a real source: BBC and others emit a grey
			// placeholder img alongside the lazy-loaded real one.
			if img == nil || (imageSrc(img) == "" && imageSrc(c) != "") {
				img = c
			}

		case atom.Figcaption:
			if figcaption == nil {
				figcaption = c
			}
		}
	}

	// A figure with no img but visible text beyond its caption is prose in
	// figure markup — a testimonial blockquote, a captioned code listing —
	// and collapsing it to a caption label would drop that content.
	if img == nil && (figcaption == nil || hasProseOutsideCaption(n)) {
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

	src := ""
	if img != nil {
		src = imageSrc(img)
	}

	p.blocks = append(p.blocks, imageBlock(caption, src, imageDisplayWidth(img)))
}

func hasProseOutsideCaption(n *html.Node) bool {
	for c := range n.ChildNodes() {
		if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
			return true
		}

		if c.Type != html.ElementNode {
			continue
		}

		if a := nodeAtom(c); a == atom.Figcaption || skippedElement(a) {
			continue
		}

		if hasProseOutsideCaption(c) {
			return true
		}
	}

	return false
}

func (p *domParser) parseTable(n *html.Node) {
	var rows [][]string

	hasHeader := false

	var visitRows func(*html.Node, bool)

	visitRows = func(group *html.Node, inHead bool) {
		for c := range group.ChildNodes() {
			if c.Type != html.ElementNode {
				continue
			}

			switch nodeAtom(c) {
			case atom.Thead:
				visitRows(c, true)

			case atom.Tbody, atom.Tfoot:
				visitRows(c, false)

			case atom.Tr:
				if row := tableRow(c); len(row) > 0 {
					if len(rows) == 0 && (inHead || allHeaderCells(c)) {
						hasHeader = true
					}

					rows = append(rows, row)
				}

			case atom.Caption:
				p.emitParagraph(collectInline(c, formatPlain, nil))
			}
		}
	}

	visitRows(n, false)

	if len(rows) > 0 {
		p.blocks = append(p.blocks, block{kind: blockTable, rows: rows, hasHeader: hasHeader})
	}
}

func allHeaderCells(tr *html.Node) bool {
	cells := 0

	for c := range tr.ChildNodes() {
		if c.Type != html.ElementNode {
			continue
		}

		switch nodeAtom(c) {
		case atom.Th:
			cells++

		case atom.Td:
			return false
		}
	}

	return cells > 0
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

	if isLatexmlNote(n) {
		return noteSpans(n, format, images)
	}

	switch nodeAtom(n) {
	case atom.Script, atom.Style, atom.Noscript, atom.Template, atom.Svg:
		return nil

	case atom.Br:
		return []span{{text: "\n", format: format}}

	case atom.Img:
		if tex := mathFallbackTeX(n); tex != "" {
			return []span{{text: latexToUnicode(tex), format: format}}
		}

		if images != nil {
			*images = append(*images, imageBlock(altText(n), imageSrc(n), imageDisplayWidth(n)))
		}

		return nil

	case atom.B, atom.Strong:
		if format == formatPlain {
			format = formatBold
		}

		return collectInline(n, format, images)

	case atom.Em, atom.I, atom.Var, atom.Dfn:
		if format == formatPlain {
			format = formatItalic
		}

		return collectInline(n, format, images)

	case atom.U, atom.Ins:
		if format == formatPlain {
			format = formatUnderline
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

	case atom.Math:
		return mathSpans(n, format)

	case atom.P, atom.Div, atom.Li:
		spans := []span{{text: " ", format: format}}
		spans = append(spans, collectInline(n, format, images)...)

		return append(spans, span{text: " ", format: format})

	default:
		return collectInline(n, format, images)
	}
}

// linkSpans unwraps an anchor to its text and keeps the target on each span,
// for rendering as a styled OSC 8 hyperlink.
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

// Unicode has no superscript q and is missing most subscript consonants;
// scriptSpans and scriptify fall back to plain forms when a rune is absent.
var superscriptRunes = map[rune]rune{
	'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴',
	'5': '⁵', '6': '⁶', '7': '⁷', '8': '⁸', '9': '⁹',
	'+': '⁺', '-': '⁻', '=': '⁼', '(': '⁽', ')': '⁾',
	'a': 'ᵃ', 'b': 'ᵇ', 'c': 'ᶜ', 'd': 'ᵈ', 'e': 'ᵉ', 'f': 'ᶠ', 'g': 'ᵍ',
	'h': 'ʰ', 'i': 'ⁱ', 'j': 'ʲ', 'k': 'ᵏ', 'l': 'ˡ', 'm': 'ᵐ', 'n': 'ⁿ',
	'o': 'ᵒ', 'p': 'ᵖ', 'r': 'ʳ', 's': 'ˢ', 't': 'ᵗ', 'u': 'ᵘ', 'v': 'ᵛ',
	'w': 'ʷ', 'x': 'ˣ', 'y': 'ʸ', 'z': 'ᶻ',
}

var subscriptRunes = map[rune]rune{
	'0': '₀', '1': '₁', '2': '₂', '3': '₃', '4': '₄',
	'5': '₅', '6': '₆', '7': '₇', '8': '₈', '9': '₉',
	'+': '₊', '-': '₋', '=': '₌', '(': '₍', ')': '₎',
	'a': 'ₐ', 'e': 'ₑ', 'h': 'ₕ', 'i': 'ᵢ', 'j': 'ⱼ', 'k': 'ₖ', 'l': 'ₗ',
	'm': 'ₘ', 'n': 'ₙ', 'o': 'ₒ', 'p': 'ₚ', 'r': 'ᵣ', 's': 'ₛ', 't': 'ₜ',
	'u': 'ᵤ', 'v': 'ᵥ', 'x': 'ₓ',
}

func hasClass(n *html.Node, class string) bool {
	return slices.Contains(strings.Fields(attr(n, "class")), class)
}

func descendantWithClass(n *html.Node, class string) *html.Node {
	for c := range n.Descendants() {
		if c.Type == html.ElementNode && hasClass(c, class) {
			return c
		}
	}

	return nil
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
