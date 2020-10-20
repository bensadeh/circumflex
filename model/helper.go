package model

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