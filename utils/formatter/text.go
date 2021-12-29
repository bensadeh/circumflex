package formatter

import (
	"code.rocketnine.space/tslocum/cview"
	text "github.com/MichaelMure/go-term-text"
)

const (
	resetStyle                   = "[::-]"
	resetForeground              = "[-::]"
	resetForegroundAndBackground = "[-:-:]"
)

func Red(text string) string {
	return "[maroon]" + text + resetForeground
}

func Green(text string) string {
	return "[green]" + text + resetForeground
}

func Yellow(text string) string {
	return "[olive]" + text + resetForeground
}

func Blue(text string) string {
	return "[navy]" + text + resetForeground
}

func Magenta(text string) string {
	return "[purple]" + text + resetForeground
}

func Cyan(text string) string {
	return "[teal]" + text + resetForeground
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

func Year(text string) string {
	return cview.TranslateANSI("\u001b[48;5;238m") +
		cview.TranslateANSI("\u001B[38;5;3m") +
		text + resetForegroundAndBackground
}

func BlackOnOrange(text string) string {
	return "[#0c0c0c:#FFA500]" + text + resetForegroundAndBackground
}

func BlackOnGreen(text string) string {
	return "[#0c0c0c:green]" + text + resetForegroundAndBackground
}

func BlackOnRed(text string) string {
	return "[#0c0c0c:maroon]" + text + resetForegroundAndBackground
}

func BlackOnYellow(text string) string {
	return "[#0c0c0c:olive]" + text + resetForegroundAndBackground
}

func BlackOnBlue(text string) string {
	return "[#0c0c0c:navy]" + text + resetForegroundAndBackground
}

func Len(textWithTage string) int {
	stripped := cview.StripTags([]byte(textWithTage), true, false)
	strippedString := string(stripped)

	return text.Len(strippedString)
}
