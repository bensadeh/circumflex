package cli

import (
	"clx/constants"
	"clx/less"
	"clx/settings"
	"context"
	"os"
	"os/exec"
	"strings"
)

func RunLess(ctx context.Context, content string, config *settings.Config) error {
	lesskey, err := less.NewLesskey()
	if err != nil {
		return err
	}
	defer lesskey.Remove()

	config.LesskeyPath = lesskey.Path()
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

	command := exec.CommandContext(ctx, "less", args...)

	command.Stdin = strings.NewReader(input)
	command.Stdout = os.Stdout

	return command
}
