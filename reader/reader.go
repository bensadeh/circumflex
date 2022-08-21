package reader

import (
	"fmt"
	"time"

	"clx/reader/markdown/postprocessor"
	"clx/reader/markdown/terminal"

	"clx/reader/markdown/html"
	"clx/reader/markdown/parser"

	"github.com/go-shiori/go-readability"
)

func GetArticle(url string, title string, width int, indentationSymbol string) (string, error) {
	articleInRawHTML, httpErr := readability.FromURL(url, 5*time.Second)
	if httpErr != nil {
		return "", fmt.Errorf("could not fetch url: %w", httpErr)
	}

	articleInMarkdown, mdErr := html.ConvertToMarkdown(articleInRawHTML.Content)
	if mdErr != nil {
		return "", fmt.Errorf("could not fetch url: %w", httpErr)
	}

	markdownBlocks := parser.ConvertToMarkdownBlocks(articleInMarkdown)

	articleInTerminalFormal := terminal.ConvertToTerminalFormat(markdownBlocks, width, indentationSymbol)

	header := terminal.CreateHeader(title, url, width)

	articleInTerminalFormal = postprocessor.Process(header+articleInTerminalFormal, url)

	return articleInTerminalFormal, nil
}
