package reader

import (
	"bytes"
	"clx/ansi"
	"fmt"
	"time"

	"codeberg.org/readeck/go-readability/v2"
)

func GetArticle(url string, title string, width int, indentationSymbol string) (string, error) {
	article, httpErr := readability.FromURL(url, 6*time.Second)
	if httpErr != nil {
		return "", fmt.Errorf("could not fetch URL: %w", httpErr)
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

	articleInTerminalFormal := convertToTerminalFormat(markdownBlocks, width, indentationSymbol)

	header := createHeader(title, url, width)

	articleInTerminalFormal = processArticle(header+articleInTerminalFormal, url)

	return articleInTerminalFormal, nil
}
