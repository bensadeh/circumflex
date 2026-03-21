package reader

import (
	"bytes"
	"clx/ansi"
	"context"
	"fmt"
	"net/http"
	nurl "net/url"

	"codeberg.org/readeck/go-readability/v2"
)

func GetArticle(ctx context.Context, url string, title string, width int, indentationSymbol string) (string, error) {
	parsedURL, err := nurl.ParseRequestURI(url)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		return "", fmt.Errorf("could not fetch URL: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

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

	header := createHeader(title, url, width)

	articleInTerminalFormal = processArticle(header+articleInTerminalFormal, url)

	return articleInTerminalFormal, nil
}
