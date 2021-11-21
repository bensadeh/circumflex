package reader

import (
	"clx/markdown"
	"clx/markdown/preprocessor"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown/plugin"

	"github.com/PuerkitoBio/goquery"

	md "github.com/JohannesKaufmann/html-to-markdown"

	"github.com/go-shiori/go-readability"
)

func GetNew(url string) (string, error) {
	art, httpErr := fetch(url)
	if httpErr != nil {
		return "", fmt.Errorf("could not fetch url: %w", httpErr)
	}

	href := md.Rule{
		Filter: []string{"a"},
		Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
			// If the span element has not the classname `bb_strike` return nil.
			// That way the next rules will apply. In this case the commonmark rules.
			// -> return nil -> next rule applies
			//if !selec.HasClass("href") {
			//	return nil
			//}

			// Trim spaces so that the following does NOT happen: `~ and cake~`.
			// Because of the space it is not recognized as strikethrough.
			// -> trim spaces at begin&end of string when inside strong/italic/...
			content = strings.TrimSpace(content)
			// return md.String("[" + content + "]")
			return md.String(content)
		},
	}

	italic := md.Rule{
		Filter: []string{"i"},
		Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
			// If the span element has not the classname `bb_strike` return nil.
			// That way the next rules will apply. In this case the commonmark rules.
			// -> return nil -> next rule applies
			//if !selec.HasClass("href") {
			//	return nil
			//}

			// Trim spaces so that the following does NOT happen: `~ and cake~`.
			// Because of the space it is not recognized as strikethrough.
			// -> trim spaces at begin&end of string when inside strong/italic/...
			content = strings.TrimSpace(content)
			return md.String(markdown.ItalicStart + content + markdown.ItalicStop)
		},
	}

	opt := &md.Options{}
	converter := md.NewConverter("", true, opt)
	converter.AddRules(href)
	converter.AddRules(italic)
	converter.Use(plugin.Table())
	// converter.AddRules(span)

	art.Content = preprocessor.ConvertItalicTags(art.Content)
	art.Content = preprocessor.ConvertBoldTags(art.Content)

	markdown, err := converter.ConvertString(art.Content)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("md ->", markdown)

	markdown = strings.ReplaceAll(markdown, "<span>", "")
	markdown = strings.ReplaceAll(markdown, "</span>", "")

	return markdown, nil
}

func fetch(rawURL string) (readability.Article, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Get(rawURL)
	if err != nil {
		return readability.Article{}, fmt.Errorf("could not fetch rawURL: %w", err)
	}

	defer response.Body.Close()

	pageURL, urlErr := url.Parse(rawURL)
	if urlErr != nil {
		panic(urlErr)
	}

	art, readabilityErr := readability.FromReader(response.Body, pageURL)
	if readabilityErr != nil {
		return readability.Article{}, fmt.Errorf("could not fetch rawURL: %w", readabilityErr)
	}

	return art, nil
}
