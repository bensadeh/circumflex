package comment

import (
	"regexp"
	"strings"
)

var (
	// HN usernames are letters, digits, _ and -; a trailing period is
	// sentence punctuation, not part of the handle.
	reMention  = regexp.MustCompile(`((?:^| )\B@[\w-]+)`)
	reVariable = regexp.MustCompile(`(\$+[a-zA-Z_\-]+)`)
	reYCLabel  = regexp.MustCompile(`\((YC [SWFXP]\d{2})\)`)
	reURL      = regexp.MustCompile(`https?://([^,"\) \n]+)`)
)

// tokenizeVerbatim carves out the content that must survive later passes
// untouched: inline code and bare URLs. It runs before typography, so an
// ellipsis or double dash inside either stays exactly as written.
func tokenizeVerbatim(b *Block) {
	switch b.kind {
	case blockParagraph:
		tokenizeBacktickPairs(b)
		splitText(b, linkifyURLs)
	case blockQuote:
		splitText(b, linkifyURLs)
	case blockCode, blockDeleted:
	}
}

// tokenizeProse splits the remaining text into semantic prose tokens.
// Quotes get none — a quote reads as one quiet block. Tokenizers work on
// plain and italic text — a token keeps its role either way, dropping the
// italic — and never touch spans that already carry a role, so they cannot
// corrupt each other's output.
func tokenizeProse(blocks []Block) {
	for i := range blocks {
		b := &blocks[i]
		if b.kind != blockParagraph {
			continue
		}

		splitText(b, tokenizeMentions)

		// Leftover backticks (an unpaired one, or ticks inside link text)
		// suppress variable highlighting: half-marked code reads worse with
		// dollar signs lighting up inside it.
		if !containsBacktick(b) {
			splitText(b, tokenizeVariables)
		}

		splitText(b, tokenizeAbbreviations)
		splitText(b, tokenizeReferences)
		splitText(b, tokenizeYCLabels)
	}
}

// tokenizeContext tells a tokenizer where the text it sees sits in the block.
type tokenizeContext struct {
	// atBlockStart reports whether the text opens the block bare, where a
	// start-of-string anchor may match.
	atBlockStart bool
	// base is the format non-token remainder pieces inherit.
	base spanFormat
}

// splitText rewrites each plain and italic span through fn. fn returns nil
// to keep the span unchanged; remainder pieces around the tokens it finds
// keep the source span's format.
func splitText(b *Block, fn func(text string, ctx tokenizeContext) []span) {
	out := make([]span, 0, len(b.spans))

	for i, s := range b.spans {
		if s.format != spanPlain && s.format != spanItalic {
			out = append(out, s)

			continue
		}

		ctx := tokenizeContext{
			atBlockStart: i == 0 && s.format == spanPlain,
			base:         s.format,
		}

		parts := fn(s.text, ctx)
		if parts == nil {
			out = append(out, s)

			continue
		}

		out = append(out, parts...)
	}

	b.spans = out
}

// tokenizeBacktickPairs converts `code` runs to inline-code spans, dropping
// the backticks. An odd total leaves everything untouched — half-marked code
// reads better verbatim than half-styled. The open/close state carries
// across the block's text spans.
func tokenizeBacktickPairs(b *Block) {
	total := 0

	for _, s := range b.spans {
		if s.format == spanPlain || s.format == spanItalic {
			total += strings.Count(s.text, "`")
		}
	}

	if total == 0 || total%2 != 0 {
		return
	}

	open := false

	out := make([]span, 0, len(b.spans))

	for _, s := range b.spans {
		if s.format != spanPlain && s.format != spanItalic {
			out = append(out, s)

			continue
		}

		for i, part := range strings.Split(s.text, "`") {
			if i > 0 {
				open = !open
			}

			if part == "" {
				continue
			}

			format := s.format
			if open {
				format = spanCodeInline
			}

			out = append(out, span{text: part, format: format})
		}
	}

	b.spans = out
}

func containsBacktick(b *Block) bool {
	for _, s := range b.spans {
		if strings.Contains(s.text, "`") {
			return true
		}
	}

	return false
}

// linkifyURLs marks bare URLs. A trailing period is sentence punctuation,
// not part of the URL, and stays in the prose.
func linkifyURLs(text string, ctx tokenizeContext) []span {
	locs := reURL.FindAllStringIndex(text, -1)
	if locs == nil {
		return nil
	}

	var out []span

	last := 0
	matched := false

	for _, loc := range locs {
		url := strings.TrimRight(text[loc[0]:loc[1]], ".")
		if !strings.Contains(url, "://") || strings.HasSuffix(url, "://") {
			continue
		}

		matched = true

		if loc[0] > last {
			out = append(out, span{text: text[last:loc[0]], format: ctx.base})
		}

		out = append(out, span{text: url, format: spanLink, href: url})
		last = loc[0] + len(url)
	}

	if !matched {
		return nil
	}

	if last < len(text) {
		out = append(out, span{text: text[last:], format: ctx.base})
	}

	return out
}

// tokenizeMentions marks @handles, including the leading space in the span.
// The start-of-text anchor only applies where the block opens with the
// handle itself.
func tokenizeMentions(text string, ctx tokenizeContext) []span {
	return splitByRegexp(text, ctx, reMention, func(match string, at int) *span {
		if at == 0 && !strings.HasPrefix(match, " ") && !ctx.atBlockStart {
			return nil
		}

		return &span{text: match, format: spanMention}
	})
}

func tokenizeVariables(text string, ctx tokenizeContext) []span {
	return splitByRegexp(text, ctx, reVariable, func(match string, _ int) *span {
		return &span{text: match, format: spanVariable}
	})
}

func tokenizeAbbreviations(text string, ctx tokenizeContext) []span {
	return splitByLiterals(text, ctx, []string{"IANAL", "IAAL"}, func(match string) span {
		return span{text: match, format: spanAbbreviation}
	})
}

// referenceLiterals derives from referenceStyles so the tokenizer and the
// renderer cannot drift apart on which references exist.
var referenceLiterals = func() []string {
	out := make([]string, len(referenceStyles))
	for i, r := range referenceStyles {
		out[i] = "[" + r.digits + "]"
	}

	return out
}()

func tokenizeReferences(text string, ctx tokenizeContext) []span {
	return splitByLiterals(text, ctx, referenceLiterals, func(match string) span {
		return span{text: strings.Trim(match, "[]"), format: spanReference}
	})
}

func tokenizeYCLabels(text string, ctx tokenizeContext) []span {
	return splitByRegexp(text, ctx, reYCLabel, func(match string, _ int) *span {
		return &span{text: strings.Trim(match, "()"), format: spanYCLabel}
	})
}

// splitByRegexp splits text at re's matches, mapping each match through mk.
// mk may reject a match (nil) to keep it as plain text; at is the match's
// byte offset. A text without matches returns nil so callers keep the
// original span.
func splitByRegexp(text string, ctx tokenizeContext, re *regexp.Regexp, mk func(match string, at int) *span) []span {
	idxs := re.FindAllStringIndex(text, -1)
	if idxs == nil {
		return nil
	}

	var out []span

	last := 0
	matched := false

	for _, loc := range idxs {
		tok := mk(text[loc[0]:loc[1]], loc[0])
		if tok == nil {
			continue
		}

		matched = true

		if loc[0] > last {
			out = append(out, span{text: text[last:loc[0]], format: ctx.base})
		}

		out = append(out, *tok)
		last = loc[1]
	}

	if !matched {
		return nil
	}

	if last < len(text) {
		out = append(out, span{text: text[last:], format: ctx.base})
	}

	return out
}

// splitByLiterals splits text at every occurrence of the given literal
// strings, scanning left to right.
func splitByLiterals(text string, ctx tokenizeContext, literals []string, mk func(match string) span) []span {
	var out []span

	rest := text

	for {
		best, bestLit := -1, ""

		for _, lit := range literals {
			if idx := strings.Index(rest, lit); idx >= 0 && (best < 0 || idx < best) {
				best, bestLit = idx, lit
			}
		}

		if best < 0 {
			break
		}

		if best > 0 {
			out = append(out, span{text: rest[:best], format: ctx.base})
		}

		out = append(out, mk(bestLit))
		rest = rest[best+len(bestLit):]
	}

	if out == nil {
		return nil
	}

	if rest != "" {
		out = append(out, span{text: rest, format: ctx.base})
	}

	return out
}
