package html

import (
	"strings"

	"github.com/f01c33/clx/reader/markdown"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
)

func ConvertToMarkdown(article string) (string, error) {
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

	return converter.ConvertString(article)
}
