package new_reader

import (
	"fmt"
	"strings"
	"time"

	"clx/markdown"

	"github.com/PuerkitoBio/goquery"

	"github.com/JohannesKaufmann/html-to-markdown/plugin"

	md "github.com/JohannesKaufmann/html-to-markdown"

	"github.com/go-shiori/go-readability"
)

func GetNew(url string) (string, error) {
	articleInRawHTML, httpErr := readability.FromURL(url, 5*time.Second)
	if httpErr != nil {
		return "", fmt.Errorf("could not fetch url: %w", httpErr)
	}

	articleInMarkdown, mdErr := convertHtmlToMarkdown(articleInRawHTML)
	if mdErr != nil {
		return "", fmt.Errorf("could not fetch url: %w", httpErr)
	}

	return articleInMarkdown, nil
}

func convertHtmlToMarkdown(art readability.Article) (string, error) {
	// Remove hyperlink tags
	href := md.Rule{
		Filter: []string{"a"},
		Replacement: func(content string, s *goquery.Selection, opt *md.Options) *string {
			content = strings.TrimSpace(content)

			return md.String(content)
		},
	}

	// Convert italic HTML tags to our own CLX_ITALIC tag because these are easier to work with after converting
	// the HTML page to Markdown
	italic := md.Rule{
		Filter: []string{"i", "em"},
		Replacement: func(content string, s *goquery.Selection, opt *md.Options) *string {
			content = strings.TrimSpace(content)

			return md.String(markdown.ItalicStart + content + markdown.ItalicStop)
		},
	}

	bold := md.Rule{
		Filter: []string{"b", "strong"},
		Replacement: func(content string, s *goquery.Selection, opt *md.Options) *string {
			content = strings.TrimSpace(content)

			return md.String(markdown.BoldStart + content + markdown.BoldStop)
		},
	}

	converter := md.NewConverter("", true, &md.Options{})
	converter.AddRules(href)
	converter.AddRules(italic)
	converter.AddRules(bold)
	converter.Use(plugin.Table())

	return converter.ConvertString(art.Content)
}
