package browser

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func Open(url string) {
	if browser := os.Getenv("CLX_BROWSER"); browser != "" {
		commandAndArgs := strings.Fields(browser)
		command := commandAndArgs[0]
		args := append(commandAndArgs[1:], url)

		cmd := exec.Command(command, args...)
		_ = cmd.Start()

		return
	}

	if browser := os.Getenv("BROWSER"); browser != "" {
		commandAndArgs := strings.Fields(browser)
		command := commandAndArgs[0]
		args := append(commandAndArgs[1:], url)

		cmd := exec.Command(command, args...)
		_ = cmd.Start()

		return
	}

	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("xdg-open", url)
		_ = cmd.Start()

	case "windows":
		cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		_ = cmd.Start()

	case "darwin":
		cmd := exec.Command("open", url)
		_ = cmd.Start()

	default:
		panic("unsupported platform")
	}
}
