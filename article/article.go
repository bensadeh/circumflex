package article

import (
	"context"
	"fmt"
	nurl "net/url"
	"strings"

	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/scrollbar"
	"github.com/bensadeh/circumflex/style"

	"golang.org/x/net/html"
)

// Parsed holds the block representation of a fetched article, allowing
// re-rendering at different widths without re-fetching.
type Parsed struct {
	blocks []block

	// Title is the page's own title as readability extracted it; empty for
	// plain-text pages. Articles opened from a story use the story title
	// instead — this one names pages reached by following links.
	Title string
}

// Parse fetches the article at url and turns it into renderable blocks,
// downloading and decoding its images so reader mode can display them. Whether
// the images are actually shown is decided later, at render time.
func Parse(ctx context.Context, url string) (*Parsed, error) {
	parsedURL, err := nurl.ParseRequestURI(url)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	body, contentType, parsedURL, err := fetchArticle(ctx, url, parsedURL)
	if err != nil {
		return nil, err
	}

	var (
		blocks []block
		title  string
	)

	switch {
	case isMarkdown(contentType, parsedURL, body):
		blocks, title, err = parseMarkdownBlocks(body, parsedURL)
		if err != nil {
			return nil, err
		}

	case isPlainText(contentType, body):
		blocks = parseTextBlocks(string(body))

	default:
		node, pageTitle, err := extractReadable(body, parsedURL)
		if err != nil {
			return nil, err
		}

		title = pageTitle
		blocks = parseBlocks(node)

		if usesMathRenderer(body) {
			convertMath(blocks)
		}

		blocks = restoreLeadImage(body, blocks)
	}

	blocks = applySiteRules(blocks, parsedURL.Hostname())

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no readable content found at %s", parsedURL.Hostname())
	}

	fetchImages(ctx, blocks, parsedURL)

	return &Parsed{blocks: blocks, Title: title}, nil
}

// NewParsedFromHTML skips fetching and readability extraction; for tests.
func NewParsedFromHTML(src string) *Parsed {
	node, err := html.Parse(strings.NewReader(src))
	if err != nil {
		return &Parsed{}
	}

	return &Parsed{blocks: parseBlocks(node)}
}

// HasImages reports whether any block holds decoded image pixels the h/l
// toggle can reveal. Figures count only when the terminal composites Kitty
// graphics — below that tier they render their description either way.
func (p *Parsed) HasImages(kitty bool) bool {
	for i := range p.blocks {
		b := &p.blocks[i]
		if b.kind != blockImage || b.img == nil {
			continue
		}

		if !b.figure || (kitty && b.kitty != nil) {
			return true
		}
	}

	return false
}

// Rendered is one rendering of a Parsed article at a particular width.
type Rendered struct {
	Body string

	// BlockStarts holds the line index each block starts on, so a scroll
	// position can be re-anchored to the same block after a re-render
	// changes block heights (image toggling, resizing).
	BlockStarts []int

	// HeadingStarts holds the line index of each section heading, for
	// jumping between sections.
	HeadingStarts []int
}

// RenderWithHeader wraps prose at contentWidth; code boxes span at least
// contentWidth and grow toward screenWidth, which verbatim text and tables
// extend to directly. A screenWidth of 0 keeps everything at contentWidth.
// The right edge reserves the scrollbar column so a full-width line is not
// clipped by the bar. images controls whether decoded images render as art
// or fall back to a text label.
func (p *Parsed) RenderWithHeader(contentWidth, screenWidth int, header string, images ImageOptions) Rendered {
	margin := strings.Repeat(" ", layout.ReaderViewLeftMargin)

	codeWidth := contentWidth
	if screenWidth > 0 {
		codeWidth = max(contentWidth, screenWidth-layout.ReaderViewLeftMargin-scrollbar.Width)
	}

	parts := renderParts(p.blocks, contentWidth, codeWidth, images)
	starts := blockStarts(parts, strings.Count(header, "\n"))

	var headings []int

	for i, part := range parts {
		if part.kind == blockHeading {
			headings = append(headings, starts[i])
		}
	}

	return Rendered{
		Body:          header + style.PrefixLines(joinParts(parts), margin),
		BlockStarts:   starts,
		HeadingStarts: headings,
	}
}
