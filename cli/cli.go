package cli

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"clx/constants/unicode"
	"clx/settings"
)

func Less(input string, config *settings.Config) *exec.Cmd {
	args := []string{
		"--RAW-CONTROL-CHARS",
		"--pattern=" + unicode.InvisibleCharacterForTopLevelComments,
		"--ignore-case",
		"--lesskey-src=" + config.LesskeyPath,
		"--tilde",
		"--use-color",
		"-P?e" + "\u001B[38;5;5m" + "E" + "\u001B[38;5;3m" + "n" + "\u001B[38;5;4m" + "d " + "\033[0m",
		"-DSy",
		"-DP-",
	}

	if config.AutoExpandComments {
		args = append(args, "+&!"+unicode.InvisibleCharacterForCollapse)
	} else {
		args = append(args, "+&!"+unicode.InvisibleCharacterForExpansion)
	}

	command := exec.Command("less", args...)

	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	return command
}

func EnableNerdFontsInLess() {
	os.Setenv("LESSUTFCHARDEF", "E000-F8FF:p,F0000-FFFFD:p,100000-10FFFD:p")
}

func ClearScreen() {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}

func VerifyLessVersion(minimumVersion int) (isValid bool, currentVersion string) {
	lessVersionInfo := getLessVersionInfo()

	lessVersionInfoWords := strings.Fields(lessVersionInfo)
	if len(lessVersionInfoWords) < 1 {
		panic("Could not parse less version info")
	}

	lessVersion, err := strconv.ParseFloat(lessVersionInfoWords[1], 64)
	if err != nil {
		panic(err)
	}

	return int(lessVersion) >= minimumVersion, lessVersionInfoWords[1]
}

func getLessVersionInfo() string {
	command := exec.Command("less", "--version")

	output, commandError := command.Output()
	if commandError != nil {
		panic(commandError)
	}

	return string(output)
}
