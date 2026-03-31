package article

import (
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
)

func convertToMarkdown(article string) (string, error) {
	href := md.Rule{
		Filter: []string{"a"},
		Replacement: func(content string, s *goquery.Selection, opt *md.Options) *string {
			content = strings.TrimSpace(content)

			return new(content)
		},
	}

	italic := md.Rule{
		Filter: []string{"i", "em"},
		Replacement: func(content string, s *goquery.Selection, opt *md.Options) *string {
			content = strings.TrimSpace(content)

			return new(italicStart + content + italicStop)
		},
	}

	bold := md.Rule{
		Filter: []string{"b", "strong"},
		Replacement: func(content string, s *goquery.Selection, opt *md.Options) *string {
			content = strings.TrimSpace(content)

			return &content
		},
	}

	converter := md.NewConverter("", true, &md.Options{})
	converter.AddRules(href)
	converter.AddRules(italic)
	converter.AddRules(bold)
	converter.Use(plugin.Table())

	return converter.ConvertString(article)
}
