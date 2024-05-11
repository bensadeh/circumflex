package reader

import (
	"fmt"
	"time"

	ansi "github.com/f01c33/clx/utils/strip-ansi"

	"github.com/f01c33/clx/reader/markdown/postprocessor"
	"github.com/f01c33/clx/reader/markdown/terminal"

	"github.com/f01c33/clx/reader/markdown/html"
	"github.com/f01c33/clx/reader/markdown/parser"

	"github.com/go-shiori/go-readability"
)

func GetArticle(url string, title string, width int, indentationSymbol string) (string, error) {
	articleInRawHtml, httpErr := readability.FromURL(url, 6*time.Second)
	if httpErr != nil {
		return "", fmt.Errorf("could not fetch url: %w", httpErr)
	}

	articleContentInRawHtmlAndSanitized := ansi.Strip(articleInRawHtml.Content)

	articleInMarkdown, mdErr := html.ConvertToMarkdown(articleContentInRawHtmlAndSanitized)
	if mdErr != nil {
		return "", fmt.Errorf("could not fetch url: %w", httpErr)
	}

	markdownBlocks := parser.ConvertToMarkdownBlocks(articleInMarkdown)

	articleInTerminalFormal := terminal.ConvertToTerminalFormat(markdownBlocks, width, indentationSymbol)

	header := terminal.CreateHeader(title, url, width)

	articleInTerminalFormal = postprocessor.Process(header+articleInTerminalFormal, url)

	return articleInTerminalFormal, nil
}
