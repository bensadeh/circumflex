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
	commandReferences := exec.Command("lynx", "-stdin", "-display_charset=utf-8", "-assume_charset=utf-8",
		"-listonly", "-nomargins", "-nonumbers", "-hiddenlinks=ignore", "-notitle", "-dump")
	commandReferences.Stdin = strings.NewReader(input)

	ref, errR := commandReferences.CombinedOutput()
	if errR != nil {
		panic(errR)
	}

	references := string(ref)

	numberOfReferences := len(strings.Split(references, newLine))
	isManyReferences := numberOfReferences > 16
	articleArguments := []string{
		"-stdin", "-display_charset=utf-8", "-assume_charset=utf-8", "-dump",
		"-hiddenlinks=ignore",
	}

	if isManyReferences {
		references = ""

		articleArguments = append(articleArguments, "-nonumbers")
	}

	commandArticle := exec.Command("lynx", articleArguments...)
	commandArticle.Stdin = strings.NewReader(input)

	art, errA := commandArticle.CombinedOutput()
	if errA != nil {
		panic(errA)
	}

	article := string(art)

	return article, references
}

func ClearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}
