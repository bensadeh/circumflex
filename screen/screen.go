package screen

import (
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func GetTerminalHeight() int {
	height, err := terminal.Height()
	if err != nil {
		panic("Could not determine terminal height")
	}

	return int(height)
}

func GetTerminalWidth() int {
	width, err := terminal.Width()
	if err != nil {
		panic("Could not determine terminal width")
	}

	return int(width)
}

func GetSubmissionsToShow(screenHeight int, maxStories int) int {
	topBarHeight := 2
	footerHeight := 2
	adjustedHeight := screenHeight - topBarHeight - footerHeight

	return min(adjustedHeight/2, maxStories)
}

func min(x, y int) int {
	if x > y {
		return y
	}

	return x
}
