package cli

import (
	"os"
	"os/exec"
	"strings"

	"clx/constants/unicode"
)

func Less(input string, pathToLesskey string) *exec.Cmd {
	command := exec.Command("less",
		"--RAW-CONTROL-CHARS",
		"--pattern="+unicode.ZeroWidthSpace,
		"--ignore-case",
		"--lesskey-src="+pathToLesskey,
		"--tilde",
		"--use-color",
		"-P?e"+"\u001B[48;5;232m "+"\u001B[38;5;200m"+"E"+"\u001B[38;5;214m"+"n"+"\u001B[38;5;69m"+"d "+"\033[0m",
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
