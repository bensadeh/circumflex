package cli

import (
	"clx/constants"
	"clx/less"
	"clx/settings"
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func RunLess(ctx context.Context, content string, config *settings.Config) error {
	lesskey, err := less.NewLesskey()
	if err != nil {
		return err
	}
	defer lesskey.Remove()

	config.LesskeyPath = lesskey.GetPath()
	command := Less(ctx, content, config)

	return command.Run()
}

func Less(ctx context.Context, input string, config *settings.Config) *exec.Cmd {
	args := []string{
		"--RAW-CONTROL-CHARS",
		"--pattern=" + constants.InvisibleCharacterForTopLevelComments,
		"--ignore-case",
		"--lesskey-src=" + config.LesskeyPath,
		"--tilde",
		"--use-color",
		"-DSy",
		"-DP-",
	}

	if config.AutoExpandComments {
		args = append(args, "+&!"+constants.InvisibleCharacterForCollapse)
	} else {
		args = append(args, "+&!"+constants.InvisibleCharacterForExpansion)
	}

	command := exec.CommandContext(ctx, "less", args...)

	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	return command
}

func EnableNerdFontsInLess() {
	_ = os.Setenv("LESSUTFCHARDEF", "E000-F8FF:p,F0000-FFFFD:p,100000-10FFFD:p")
}

func ClearScreen(ctx context.Context) {
	c := exec.CommandContext(ctx, "clear")
	c.Stdout = os.Stdout
	_ = c.Run()
}

func VerifyLessVersion(minimumVersion int) (isValid bool, currentVersion string) {
	lessVersionInfo := getLessVersionInfo()

	lessVersionInfoWords := strings.Fields(lessVersionInfo)
	if len(lessVersionInfoWords) < 1 {
		return false, ""
	}

	lessVersion, err := strconv.ParseFloat(lessVersionInfoWords[1], 64)
	if err != nil {
		return false, ""
	}

	isValid = int(lessVersion) >= minimumVersion
	currentVersion = lessVersionInfoWords[1]

	return isValid, currentVersion
}

func getLessVersionInfo() string {
	command := exec.CommandContext(context.Background(), "less", "--version")

	output, commandError := command.Output()
	if commandError != nil {
		panic(commandError)
	}

	return string(output)
}
