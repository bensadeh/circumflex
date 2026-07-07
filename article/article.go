package article

import (
	"context"
	"fmt"
	nurl "net/url"
	"strings"

	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/style"

	"golang.org/x/net/html"
)

// Parsed holds the block representation of a fetched article, allowing
// re-rendering at different widths without re-fetching.
type Parsed struct {
	blocks []block
}

// Parse fetches an article and converts it to the intermediate block
// representation: readability extracts the readable DOM, the walker maps it
// to blocks, and per-site rules strip known boilerplate.
func Parse(ctx context.Context, url string) (*Parsed, error) {
	parsedURL, err := nurl.ParseRequestURI(url)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	node, err := fetchDocument(ctx, url, parsedURL)
	if err != nil {
		return nil, err
	}

	blocks := parseBlocks(node)
	blocks = applySiteRules(blocks, parsedURL.Hostname())

	return &Parsed{blocks: blocks}, nil
}

// NewParsedFromHTML creates a Parsed value from an HTML fragment, bypassing
// the network fetch and readability. Intended for tests.
func NewParsedFromHTML(src string) *Parsed {
	node, err := html.Parse(strings.NewReader(src))
	if err != nil {
		return &Parsed{}
	}

	return &Parsed{blocks: parseBlocks(node)}
}

func (p *Parsed) RenderWithHeader(width int, header string) string {
	margin := strings.Repeat(" ", layout.ReaderViewLeftMargin)

	return header + style.PrefixLines(renderBlocks(p.blocks, width), margin)
}
