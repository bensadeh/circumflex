package article

import (
	"bytes"
	"context"
	"fmt"
	nurl "net/url"
	"time"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/version"

	"codeberg.org/readeck/go-readability/v2"
	"resty.dev/v3"
)

const (
	fetchTimeout = 4 * time.Second
	retryCount   = 2
)

// discardLogger silences resty's internal logging so that WARN/ERROR
// messages on context cancellation don't corrupt the TUI.
type discardLogger struct{}

func (discardLogger) Errorf(string, ...any) {}
func (discardLogger) Warnf(string, ...any)  {}
func (discardLogger) Debugf(string, ...any) {}

// Parsed holds the intermediate representation of a fetched article,
// allowing re-rendering at different widths without re-fetching.
type Parsed struct {
	blocks []*block
	url    string
}

// Render formats the parsed article for terminal display at the given width,
// using the default meta header.
func (p *Parsed) Render(width int) string {
	return p.RenderWithHeader(width, createHeader(p.url, width))
}

// RenderWithHeader formats the parsed article with a custom header block
// prepended before postprocessing.
func (p *Parsed) RenderWithHeader(width int, header string) string {
	content := convertToTerminalFormat(p.blocks, width)

	return header + processArticle(content, p.url, width)
}

// Parse fetches and parses an article, returning the intermediate
// representation that can be rendered at any width.
func Parse(ctx context.Context, url string) (*Parsed, error) {
	parsedURL, err := nurl.ParseRequestURI(url)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	client := resty.New()

	defer func() { _ = client.Close() }()

	client.SetTimeout(fetchTimeout)
	client.SetRetryCount(retryCount)
	client.SetHeader("User-Agent", version.Name+"/"+version.Version)
	client.SetLogger(discardLogger{})

	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		return nil, fmt.Errorf("could not fetch URL: %w", err)
	}

	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("server returned status %d for %s", resp.StatusCode(), parsedURL.Host)
	}

	a, err := readability.FromReader(bytes.NewReader(resp.Bytes()), parsedURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse article from %s", parsedURL.Host)
	}

	var buf bytes.Buffer
	if err := a.RenderHTML(&buf); err != nil {
		return nil, fmt.Errorf("could not extract readable content from %s", parsedURL.Host)
	}

	raw := ansi.Strip(buf.String())

	md, mdErr := convertToMarkdown(raw)
	if mdErr != nil {
		return nil, fmt.Errorf("could not convert to Markdown: %w", mdErr)
	}

	blocks := convertToMarkdownBlocks(md)
	normalizeHeaders(blocks)

	return &Parsed{blocks: blocks, url: url}, nil
}

// NewParsedFromMarkdown creates a Parsed value from raw Markdown text,
// bypassing the network fetch. Intended for tests that need a Parsed
// without hitting a real URL.
func NewParsedFromMarkdown(md string) *Parsed {
	blocks := convertToMarkdownBlocks(md)
	normalizeHeaders(blocks)

	return &Parsed{blocks: blocks, url: "https://example.com"}
}

// Fetch fetches, parses, and renders an article in one step.
// Convenience wrapper used by standalone commands.
func Fetch(ctx context.Context, url string, width int) (string, error) {
	p, err := Parse(ctx, url)
	if err != nil {
		return "", err
	}

	return p.Render(width), nil
}
