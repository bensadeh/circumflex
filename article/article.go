package article

import (
	"context"
	"fmt"
	"image"
	nurl "net/url"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/highlight"
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

// Parse fetches the article at url and turns it into renderable blocks. The
// page's images are downloaded and decoded only when images is true: nothing
// below a Kitty-graphics terminal can draw them, and there the whole
// fetch-decode-rasterize-encode pass would be discarded for a text label.
// Whether a drawable image is actually shown is still decided later, at render
// time, by the h/l toggle.
//
// Callers inside a started program pass graphics.Enabled(). The standalone
// commands cannot: they parse before the program exists, and a terminal
// without graphics support answers the probe with silence rather than a no —
// so "not supported" and "has not answered yet" are the same reading. They
// pass true and keep fetching unconditionally.
func Parse(ctx context.Context, url string, images bool) (*Parsed, error) {
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
	guessCodeLangs(blocks)

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no readable content found at %s", parsedURL.Hostname())
	}

	if images {
		fetchImages(ctx, blocks, parsedURL)
	}

	// Block text is stripped as it parses; the title arrives on its own path
	// out of readability, still carrying whatever the page put in <title>.
	// Fields also folds newlines: the title heads the view as a single line.
	title = strings.Join(strings.Fields(ansi.Strip(title)), " ")

	return &Parsed{blocks: blocks, Title: title}, nil
}

// guessCodeLangs fills in the language of unlabeled code blocks from
// highlight's structural detection; declared languages are never
// second-guessed.
func guessCodeLangs(blocks []block) {
	for i := range blocks {
		b := &blocks[i]
		if b.kind == blockCode && b.lang == "" {
			b.lang = highlight.GuessLang(b.text)
		}
	}
}

// NewParsedFromHTML skips fetching and readability extraction; for tests.
func NewParsedFromHTML(src string) *Parsed {
	node, err := html.Parse(strings.NewReader(src))
	if err != nil {
		return &Parsed{}
	}

	return &Parsed{blocks: parseBlocks(node)}
}

// HasImages reports whether any block holds pixels the h/l toggle can reveal.
// Only a terminal speaking the Kitty graphics protocol can draw them, so
// without it there is nothing to toggle and every image stays a label.
func (p *Parsed) HasImages(kitty bool) bool {
	if !kitty {
		return false
	}

	for i := range p.blocks {
		b := &p.blocks[i]
		if b.kind == blockImage && b.kitty != nil && b.imgSize != (image.Point{}) {
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
