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

func Lynx(input string) string {
	command := exec.Command("lynx", "-stdin", "-display_charset=utf-8", "-assume_charset=utf-8", "-dump")
	command.Stdin = strings.NewReader(input)

	out, err := command.CombinedOutput()
	if err != nil {
		panic(err)
	}

	return string(out)
}

func ClearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}
