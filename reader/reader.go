package reader

import (
	"clx/article"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

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

func fetch(url string) (readability.Article, error) {
	client := http.Client{
		Timeout: 4 * time.Second,
	}

	response, err := client.Get(url)
	if err != nil {
		return readability.Article{}, fmt.Errorf("could not fetch url: %w", err)
	}

	defer response.Body.Close()

	art, readabilityErr := readability.FromReader(response.Body, url)
	if readabilityErr != nil {
		return readability.Article{}, fmt.Errorf("could not fetch url: %w", readabilityErr)
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
