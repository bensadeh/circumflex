package cli

import (
	"os"
	"os/exec"
	"strings"

	"clx/settings"

	"clx/constants/unicode"
)

func Less(input string, config *settings.Config) *exec.Cmd {
	args := []string{
		"--RAW-CONTROL-CHARS",
		"--pattern=" + unicode.ZeroWidthSpace,
		"--ignore-case",
		"--lesskey-src=" + config.LesskeyPath,
		"--tilde",
		"--use-color",
		"-P?e" + "\u001B[48;5;232m " + "\u001B[38;5;200m" + "E" + "\u001B[38;5;214m" + "n" + "\u001B[38;5;69m" + "d " + "\033[0m",
		"-DSy",
		"-DP-",
	}

	if config.AutoExpandComments {
		args = append(args, "+A")
	} else {
		args = append(args, "+C")
	}

	command := exec.Command("less", args...)

	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	return command
}

func ClearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}
