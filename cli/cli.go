package cli

import (
	"os"
	"os/exec"
	"strings"
)

const (
	newLine = "\n"
)

func Less(input string) {
	command := exec.Command("less", "-r")
	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	if err := command.Run(); err != nil {
		panic(err)
	}
}

func ParseWithLynx(input string) (string, string) {
	references := getReferences(input)

	numberOfReferences := len(strings.Split(references, newLine))
	isManyReferences := numberOfReferences > 16

	additionalArgument := ""

	if isManyReferences {
		references = ""
		additionalArgument = "-nolist"
	}

	article := getArticle(input, additionalArgument)

	return article, references
}

func getReferences(input string) string {
	command := exec.Command("lynx", "-stdin", "-display_charset=utf-8", "-assume_charset=utf-8",
		"-listonly", "-nomargins", "-nonumbers", "-hiddenlinks=ignore", "-notitle", "-unique_urls", "-dump")
	command.Stdin = strings.NewReader(input)

	references, err := command.Output()
	if err != nil {
		panic(err)
	}

	return string(references)
}

func getArticle(input string, additionalArgument string) string {
	articleArguments := []string{
		"-stdin", "-display_charset=utf-8", "-assume_charset=utf-8", "-dump",
		"-hiddenlinks=ignore", "-unique_urls",
	}

	articleArguments = append(articleArguments, additionalArgument)
	command := exec.Command("lynx", articleArguments...)
	command.Stdin = strings.NewReader(input)

	article, err := command.Output()
	if err != nil {
		panic(err)
	}

	return string(article)
}

func ClearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}
