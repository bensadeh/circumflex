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

func GetOffsetForLeftAlignedTextBlock(textWidth int) int {
	screenWidth := GetTerminalWidth()
	halfScreenWidth := screenWidth / 2
	halfTextWidth := textWidth / 2

	return screenWidth - halfScreenWidth - halfTextWidth
}

func GetOffsetToCenterText(textHeight int) int {
	screenHeight := GetTerminalHeight()
	halfScreenHeight := screenHeight / 2
	halfTextHeight := textHeight / 2

	return screenHeight - halfScreenHeight - halfTextHeight
}
