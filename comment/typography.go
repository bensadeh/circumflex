package comment

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// applyTypography rewrites plain and italic span text with typographic
// niceties: real ellipses and em-dashes, unicode fractions, CO₂, smileys.
// Inline code and URLs were tokenized out before this pass — they are
// quoted verbatim.
func applyTypography(b *Block) {
	if b.kind != blockParagraph && b.kind != blockQuote {
		return
	}

	for i := range b.spans {
		s := &b.spans[i]
		if s.format != spanPlain && s.format != spanItalic {
			continue
		}

		s.text = replaceSymbols(s.text)
	}

	convertSmileys(b)
}

var (
	reDoubleDash = regexp.MustCompile(`([a-zA-Z])--([a-zA-Z])`)
	reCO2        = regexp.MustCompile(`\bCO2\b`)
)

func replaceSymbols(text string) string {
	text = strings.ReplaceAll(text, "...", "…")
	text = reCO2.ReplaceAllString(text, "CO₂")

	text = strings.ReplaceAll(text, " -- ", " — ")
	text = reDoubleDash.ReplaceAllString(text, `$1—$2`)

	return replaceFractions(text)
}

// reFraction finds a known fraction opening a word — text start, a space or
// a bracket before it — with an optional ordinal suffix. The digit that may
// follow is checked separately: a regexp can't refuse it without consuming
// the character a neighboring fraction needs as its own boundary.
var reFraction = regexp.MustCompile(`(?:^|[ (])(1/10|1/2|1/3|2/3|1/4|3/4|1/5|2/5|3/5|4/5|1/6)(?:th)?`)

// The narrow ⅒ glyph gets a trailing space to preserve alignment; the space
// collapser folds it away when a space already follows.
var fractionGlyphs = map[string]string{
	"1/2": "½", "1/3": "⅓", "2/3": "⅔",
	"1/4": "¼", "3/4": "¾",
	"1/5": "⅕", "2/5": "⅖", "3/5": "⅗", "4/5": "⅘",
	"1/6": "⅙", "1/10": "⅒ ",
}

// replaceFractions converts fractions that end at a boundary: a following
// digit or slash keeps "1/2022", "1/25" and "1/6/2021" a date or a ratio.
func replaceFractions(text string) string {
	matches := reFraction.FindAllStringSubmatchIndex(text, -1)
	if matches == nil {
		return text
	}

	var sb strings.Builder

	last := 0

	for _, m := range matches {
		end := m[1]
		if end < len(text) && (text[end] == '/' || (text[end] >= '0' && text[end] <= '9')) {
			continue
		}

		frac := text[m[2]:m[3]]

		sb.WriteString(text[last:m[2]]) // up to and including the boundary
		sb.WriteString(fractionGlyphs[frac])
		sb.WriteString(text[m[3]:end]) // the ordinal suffix, if any

		last = end
	}

	sb.WriteString(text[last:])

	return sb.String()
}

var smileys = []struct{ from, to string }{
	{`:)`, "😊"},
	{`(:`, "😊"},
	{`:-)`, "😊"},
	{`:D`, "😄"},
	{`=)`, "😃"},
	{`=D`, "😃"},
	{`;)`, "😉"},
	{`;-)`, "😉"},
	{`:P`, "😜"},
	{`;P`, "😜"},
	{`:o`, "😮"},
	{`:O`, "😮"},
	{`:(`, "😔"},
	{`:-(`, "😔"},
	{`:/`, "😕"},
	{`:-/`, "😕"},
	{`-_-`, "😑"},
	{`:|`, "😐"},
}

// convertSmileys replaces smileys that follow whitespace and end on a word
// boundary. The whole-text exact match applies at block granularity: a
// comment that is nothing but ":)" is a smiley, while a ":)" wrapped in
// markup or mid-paragraph is not.
func convertSmileys(b *Block) {
	for _, sm := range smileys {
		if len(b.spans) == 1 && b.spans[0].format == spanPlain && b.spans[0].text == sm.from {
			b.spans[0].text = sm.to

			continue
		}

		for i := range b.spans {
			s := &b.spans[i]
			if s.format != spanPlain && s.format != spanItalic {
				continue
			}

			s.text = replaceSmiley(s.text, sm.from, sm.to)
		}
	}
}

// replaceSmiley converts " <smiley>" occurrences not glued to a following
// letter or digit, so the ":D" in ":Dave" and the ":/" in ":/etc" stay prose.
func replaceSmiley(text, from, to string) string {
	from = " " + from

	var sb strings.Builder

	for {
		idx := strings.Index(text, from)
		if idx < 0 {
			break
		}

		end := idx + len(from)

		r, size := utf8.DecodeRuneInString(text[end:])
		if size > 0 && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			sb.WriteString(text[:end])
		} else {
			sb.WriteString(text[:idx])
			sb.WriteString(" ")
			sb.WriteString(to)
		}

		text = text[end:]
	}

	sb.WriteString(text)

	return sb.String()
}

// normalizeWhitespace joins soft line breaks, collapses space runs and trims
// block edges. It runs last, over the tokenized spans.
func normalizeWhitespace(blocks []Block) {
	for i := range blocks {
		b := &blocks[i]
		if b.kind != blockParagraph && b.kind != blockQuote {
			continue
		}

		joinNewlines(b)
		collapseSpaces(b)
		trimBlockEdges(b)
	}
}

// joinNewlines resolves raw newlines in the comment text. A single newline
// is a soft break and joins into the prose — unless the next visible
// character suggests a hand-formatted line start (bracketed footnotes, list
// markers, shell prompts), which keeps its hard break. A run of newlines is
// one hard break.
func joinNewlines(b *Block) {
	for i := range b.spans {
		s := &b.spans[i]
		if s.format != spanPlain && s.format != spanItalic {
			continue
		}

		next := ""
		if i+1 < len(b.spans) {
			next = visibleLead(&b.spans[i+1])
		}

		s.text = joinNewlinesIn(s.text, next)
	}
}

func joinNewlinesIn(text, followingLead string) string {
	var sb strings.Builder

	for len(text) > 0 {
		idx := strings.IndexByte(text, '\n')
		if idx < 0 {
			sb.WriteString(text)

			break
		}

		sb.WriteString(text[:idx])

		run := 1
		for idx+run < len(text) && text[idx+run] == '\n' {
			run++
		}

		follower := text[idx+run:]
		if follower == "" {
			follower = followingLead
		}

		switch {
		case run > 1:
			sb.WriteByte('\n')
		case joinsSoftBreak(follower):
			sb.WriteByte(' ')
		default:
			sb.WriteByte('\n')
		}

		text = text[idx+run:]
	}

	return sb.String()
}

// joinsSoftBreak reports whether text reads as a prose continuation: letters,
// digits and ordinary punctuation join, while bracket-like openers and
// symbols keep their line.
func joinsSoftBreak(text string) bool {
	r, size := utf8.DecodeRuneInString(text)
	if size == 0 {
		return false
	}

	return unicode.IsLetter(r) || unicode.IsDigit(r) ||
		r == '"' || r == ' ' || r == '-' || r == '…'
}

// visibleLead is the first character a span puts on screen: references open
// with their bracket, everything else with its own text.
func visibleLead(s *span) string {
	if s.format == spanReference {
		return "["
	}

	return s.text
}

var reSpaceRun = regexp.MustCompile(` {2,}`)

func collapseSpaces(b *Block) {
	for i := range b.spans {
		s := &b.spans[i]
		if s.format != spanPlain && s.format != spanItalic {
			continue
		}

		s.text = reSpaceRun.ReplaceAllString(s.text, " ")
	}
}

// trimBlockEdges drops the whitespace HN's formatting leaves at block
// boundaries, such as the newline after a closing </code></pre>. Token
// spans end the trim — their content is not whitespace.
func trimBlockEdges(b *Block) {
	for len(b.spans) > 0 {
		first := &b.spans[0]
		if first.format != spanPlain && first.format != spanItalic {
			break
		}

		first.text = strings.TrimLeft(first.text, " \n")
		if first.text != "" {
			break
		}

		b.spans = b.spans[1:]
	}

	for len(b.spans) > 0 {
		last := &b.spans[len(b.spans)-1]
		if last.format != spanPlain && last.format != spanItalic {
			break
		}

		last.text = strings.TrimRight(last.text, " \n")
		if last.text != "" {
			break
		}

		b.spans = b.spans[:len(b.spans)-1]
	}
}
