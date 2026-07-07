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

func Parse(ctx context.Context, url string) (*Parsed, error) {
	parsedURL, err := nurl.ParseRequestURI(url)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	body, contentType, err := fetchPage(ctx, url, parsedURL)
	if err != nil {
		return nil, err
	}

	var blocks []block

	if isPlainText(contentType, body) {
		blocks = parseTextBlocks(string(body))
	} else {
		node, err := extractReadable(body, parsedURL)
		if err != nil {
			return nil, err
		}

		blocks = parseBlocks(node)
	}

	blocks = applySiteRules(blocks, parsedURL.Hostname())

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no readable content found at %s", parsedURL.Hostname())
	}

	return &Parsed{blocks: blocks}, nil
}

// NewParsedFromHTML skips fetching and readability extraction; for tests.
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
