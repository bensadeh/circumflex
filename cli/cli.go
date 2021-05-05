package cli

import (
	"os"
	"os/exec"
	"strings"
)

func Less(input string) {
	command := exec.Command("less", "-r")
	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	if err := command.Run(); err != nil {
		panic(err)
	}
}

func Lynx(input string) (string, string) {
	commandArticle := exec.Command("lynx", "-stdin", "-display_charset=utf-8", "-assume_charset=utf-8", "-dump",
		"-hiddenlinks=ignore")
	commandArticle.Stdin = strings.NewReader(input)

	art, errA := commandArticle.CombinedOutput()
	if errA != nil {
		panic(errA)
	}

	article := string(art)

	commandReferences := exec.Command("lynx", "-stdin", "-display_charset=utf-8", "-assume_charset=utf-8",
		"-listonly", "-nomargins", "-nonumbers", "-hiddenlinks=ignore", "-notitle", "-dump")
	commandReferences.Stdin = strings.NewReader(input)

	ref, errR := commandReferences.CombinedOutput()
	if errR != nil {
		panic(errR)
	}

	references := string(ref)

	return article, references
}

func ClearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}
