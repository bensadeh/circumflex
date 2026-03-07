package reader

import (
	"clx/ansi"
	"fmt"
	"time"

	"github.com/go-shiori/go-readability"
)

func GetArticle(url string, title string, width int, indentationSymbol string) (string, error) {
	articleInRawHtml, httpErr := readability.FromURL(url, 6*time.Second)
	if httpErr != nil {
		return "", fmt.Errorf("could not fetch url: %w", httpErr)
	}

	articleContentInRawHtmlAndSanitized := ansi.Strip(articleInRawHtml.Content)

	articleInMarkdown, mdErr := convertToMarkdown(articleContentInRawHtmlAndSanitized)
	if mdErr != nil {
		return "", fmt.Errorf("could not convert to markdown: %w", mdErr)
	}

	markdownBlocks := convertToMarkdownBlocks(articleInMarkdown)

	articleInTerminalFormal := convertToTerminalFormat(markdownBlocks, width, indentationSymbol)

	header := createHeader(title, url, width)

	articleInTerminalFormal = processArticle(header+articleInTerminalFormal, url)

	return articleInTerminalFormal, nil
}
