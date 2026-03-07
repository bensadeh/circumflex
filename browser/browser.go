package browser

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func Open(ctx context.Context, url string) error {
	if browser := os.Getenv("CLX_BROWSER"); browser != "" {
		commandAndArgs := strings.Fields(browser)
		command := commandAndArgs[0]
		args := make([]string, len(commandAndArgs)-1, len(commandAndArgs))
		copy(args, commandAndArgs[1:])
		args = append(args, url)

		cmd := exec.CommandContext(ctx, command, args...)
		return cmd.Start()
	}

	switch runtime.GOOS {
	case "linux", "freebsd", "openbsd", "netbsd":
		cmd := exec.CommandContext(ctx, "xdg-open", url)
		return cmd.Start()

	case "windows":
		cmd := exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", url)
		return cmd.Start()

	case "darwin":
		cmd := exec.CommandContext(ctx, "open", url)
		return cmd.Start()

	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
