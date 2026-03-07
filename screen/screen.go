package screen

import (
	"fmt"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func GetTerminalHeight() (int, error) {
	height, err := terminal.Height()
	if err != nil {
		return 0, fmt.Errorf("could not determine terminal height: %w", err)
	}

	return int(height), nil
}

func GetTerminalWidth() (int, error) {
	width, err := terminal.Width()
	if err != nil {
		return 0, fmt.Errorf("could not determine terminal width: %w", err)
	}

	return int(width), nil
}

func GetSubmissionsToShow(screenHeight int, maxStories int) int {
	topBarHeight := 2
	footerHeight := 2
	adjustedHeight := screenHeight - topBarHeight - footerHeight

	return min(adjustedHeight/2, maxStories)
}
