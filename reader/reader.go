package reader

import (
	"bytes"
	"clx/ansi"
	"clx/version"
	"context"
	"fmt"
	nurl "net/url"
	"time"

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

func Article(ctx context.Context, url string, title string, width int, indentationSymbol string) (string, error) {
	parsedURL, err := nurl.ParseRequestURI(url)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
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
			return "", ctx.Err()
		}

		return "", fmt.Errorf("could not fetch URL: %w", err)
	}

	article, err := readability.FromReader(bytes.NewReader(resp.Bytes()), parsedURL)
	if err != nil {
		return "", fmt.Errorf("could not parse article: %w", err)
	}

	var buf bytes.Buffer
	if err := article.RenderHTML(&buf); err != nil {
		return "", fmt.Errorf("could not render article: %w", err)
	}

	articleContentInRawHtmlAndSanitized := ansi.Strip(buf.String())

	articleInMarkdown, mdErr := convertToMarkdown(articleContentInRawHtmlAndSanitized)
	if mdErr != nil {
		return "", fmt.Errorf("could not convert to Markdown: %w", mdErr)
	}

	markdownBlocks := convertToMarkdownBlocks(articleInMarkdown)
	normalizeHeaders(markdownBlocks)

	articleInTerminalFormal := convertToTerminalFormat(markdownBlocks, width, indentationSymbol)

	header := createHeader(url, width)

	articleInTerminalFormal = processArticle(header+articleInTerminalFormal, url)

	return articleInTerminalFormal, nil
}
