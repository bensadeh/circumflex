package cli

import (
	"clx/constants/unicode"
	"os"
	"os/exec"
	"strings"
)

func Less(input string) {
	command := exec.Command("less",
		"--RAW-CONTROL-CHARS",
		"--pattern="+unicode.ZeroWidthSpace,
		"--ignore-case")

	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	if err := command.Run(); err != nil {
		panic(err)
	}
}

func Glow(input string) {
	command := exec.Command("glow", "-", "-p")
	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	if err := command.Run(); err != nil {
		panic(err)
	}
}

func ClearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}
