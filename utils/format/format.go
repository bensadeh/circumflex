package format

import (
	text "github.com/MichaelMure/go-term-text"
	"gitlab.com/tslocum/cview"
)

const (
	resetStyle                   = "[::-]"
	resetForeground              = "[-::]"
	resetForegroundAndBackground = "[-:-:]"
)

func BlackNoReset(text string) string {
	return "[#0c0c0c]" + text
}

func Magenta(text string) string {
	return "[purple]" + text + resetForeground
}

func Red(text string) string {
	return "[maroon]" + text + resetForeground
}

func Blue(text string) string {
	return "[navy]" + text + resetForeground
}

func Green(text string) string {
	return "[green]" + text + resetForeground
}

func White(text string) string {
	return "[white]" + text + resetForeground
}

func Dim(text string) string {
	return "[::d]" + text + resetStyle
}

func Bold(text string) string {
	return "[::b]" + text + resetStyle
}

func Reverse(text string) string {
	return "[::r]" + text + resetStyle
}

func BlackOnOrange(text string) string {
	return "[#0c0c0c:orange]" + text + resetForegroundAndBackground
}

func BlackOnRed(text string) string {
	return "[#0c0c0c:red]" + text + resetForegroundAndBackground
}

func ResetStyle() string {
	return resetStyle
}

func Len(textWithTage string) int {
	stripped := cview.StripTags([]byte(textWithTage), true, false)
	strippedString := string(stripped)

	return text.Len(strippedString)
}
