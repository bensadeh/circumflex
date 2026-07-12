package help

import (
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
	return mainMenuText(screenWidth, sections) + newPar
}

// ReaderHelpScreen and CommentHelpScreen inherit the geometry of the view
// they overlay — the same left margin and text column — so the help panels
// and footer line up with the content and footer underneath.
func ReaderHelpScreen(leftMargin, contentWidth int, withStoryNav bool) string {
	return readerText(leftMargin, contentWidth, withStoryNav) + newPar
}

func CommentHelpScreen(leftMargin, contentWidth int, enableNerdFonts bool, withStoryNav bool) string {
	return commentText(leftMargin, contentWidth, enableNerdFonts, withStoryNav) + newPar
}
