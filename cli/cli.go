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
		"--ignore-case",
		"--tilde",
		"--use-color",
		"-P?e"+"\u001B[48;5;237m "+"\u001B[38;5;200m"+"e"+"\u001B[38;5;214m"+"n"+"\u001B[38;5;69m"+"d "+"\033[0m",
		"-DSy",
		"-DP-")

	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	if err := command.Run(); err != nil {
		panic(err)
	}
}

func WrapLess(input string) *exec.Cmd {
	command := exec.Command("less",
		"--RAW-CONTROL-CHARS",
		"--pattern="+unicode.ZeroWidthSpace,
		"--ignore-case",
		"--tilde",
		"--use-color",
		"-P?e"+"\u001B[48;5;237m "+"\u001B[38;5;200m"+"E"+"\u001B[38;5;214m"+"n"+"\u001B[38;5;69m"+"d "+"\033[0m",
		"-DSy",
		"-DP-")

	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	return command
}

func ClearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}
