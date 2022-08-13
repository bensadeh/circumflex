package cli

import (
	"os"
	"os/exec"
	"strings"

	"clx/constants/unicode"
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
		"-P?e"+"\u001B[48;5;234m "+"\u001B[38;5;200m"+"E"+"\u001B[38;5;214m"+"n"+"\u001B[38;5;69m"+"d "+"\033[0m",
		"-DSy",
		"-DP-")

	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	return command
}

func LessWithLesskey(input string, pathToLesskey string) *exec.Cmd {
	command := exec.Command("less",
		"--RAW-CONTROL-CHARS",
		"--pattern="+unicode.ZeroWidthSpace,
		"--ignore-case",
		"--lesskey-src="+pathToLesskey,
		"--tilde",
		"--use-color",
		"-P?e"+"\u001B[48;5;234m "+"\u001B[38;5;200m"+"E"+"\u001B[38;5;214m"+"n"+"\u001B[38;5;69m"+"d "+"\033[0m",
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
