package help

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/key"
)

const (
	newPar = "\n\n"
)

// FitToHeight pads or truncates content to exactly height lines.
// The returned string contains height lines joined by \n with no trailing newline.
func FitToHeight(content string, height int) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")

	if len(lines) > height {
		lines = lines[:height]
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

type Section struct {
	Title string
	Color color.Color
	Items []Item
}

type Item struct {
	Key  string
	Desc string
}

func FromBinding(b key.Binding) Item {
	if !b.Enabled() {
		return Item{}
	}

	return Item{Key: b.Help().Key, Desc: b.Help().Desc}
}

func MainMenuHelpScreen(screenWidth int, sections []Section) string {
	var sb strings.Builder

	sb.WriteString(MainMenuText(screenWidth, sections) + newPar)

	return sb.String()
}

func ReaderHelpScreen(screenWidth int) string {
	var sb strings.Builder

	sb.WriteString(ReaderText(screenWidth) + newPar)

	return sb.String()
}

func CommentHelpScreen(screenWidth int, enableNerdFonts bool) string {
	var sb strings.Builder

	sb.WriteString(CommentText(screenWidth, enableNerdFonts) + newPar)

	return sb.String()
}
