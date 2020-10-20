package screen

import (
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func GetTerminalHeight() int {
	height, _ := terminal.Height()
	return int(height)
}

func GetTerminalWidth() int {
	width, _ := terminal.Width()
	return int(width)
}

func GetViewableStoriesOnSinglePage(screenHeight int, maxStories int) int {
	return min(screenHeight/2-2, maxStories)
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}