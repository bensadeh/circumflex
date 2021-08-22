package reader

import (
	"clx/article"
	"clx/markdown/preprocessor"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown/plugin"

	"github.com/PuerkitoBio/goquery"

	md "github.com/JohannesKaufmann/html-to-markdown"

	"github.com/go-shiori/go-readability"
)

const (
	newLine = "\n"
)

func Get(url string) (string, error) {
	art, httpErr := fetch(url)
	if httpErr != nil {
		return "", fmt.Errorf("could not fetch url: %w", httpErr)
	}

	content, references := parseWithLynx(art.Content)
	parsedArticle := article.Parse(art.Title, url, content, references)

	return parsedArticle, nil
}

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

	//span := md.Rule{
	//	Filter: []string{"span"},
	//	Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
	//		// If the span element has not the classname `bb_strike` return nil.
	//		// That way the next rules will apply. In this case the commonmark rules.
	//		// -> return nil -> next rule applies
	//		//if !selec.HasClass("href") {
	//		//	return nil
	//		//}
	//
	//		// Trim spaces so that the following does NOT happen: `~ and cake~`.
	//		// Because of the space it is not recognized as strikethrough.
	//		// -> trim spaces at begin&end of string when inside strong/italic/...
	//		content = strings.TrimSpace(content)
	//		// return md.String("[" + content + "]")
	//		return md.String(content)
	//	},
	//}

	opt := &md.Options{}
	converter := md.NewConverter("", true, opt)
	converter.AddRules(href)
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

func parseWithLynx(input string) (string, string) {
	references := getReferences(input)

	numberOfReferences := len(strings.Split(references, newLine))
	isManyReferences := numberOfReferences > 16

	additionalArgument := ""

	if isManyReferences {
		references = ""
		additionalArgument = "-nolist"
	}

	content := getContent(input, additionalArgument)

	return content, references
}

func getReferences(input string) string {
	command := exec.Command("lynx", "-stdin", "-display_charset=utf-8", "-assume_charset=utf-8",
		"-listonly", "-nomargins", "-nonumbers", "-hiddenlinks=ignore", "-notitle", "-dump")
	command.Stdin = strings.NewReader(input)

	references, err := command.Output()
	if err != nil {
		panic(err)
	}

	return string(references)
}

func getContent(input string, additionalArgument string) string {
	articleArguments := []string{
		"-stdin", "-display_charset=utf-8", "-assume_charset=utf-8", "-dump",
		"-hiddenlinks=ignore",
	}

	articleArguments = append(articleArguments, additionalArgument)
	command := exec.Command("lynx", articleArguments...)
	command.Stdin = strings.NewReader(input)

	content, err := command.Output()
	if err != nil {
		panic(err)
	}

	return string(content)
}
