package help

import (
	"strings"

	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"
)

// Aligned with the default `meta` block: commentWidth (70) + border (2) outer,
// MarginLeft 1. So the panels sit in the same screen position as the meta box.
const (
	panelOuterWidth = 72
	panelLeftMargin = 1
)

func mainMenuText(screenWidth int, sections []Section) string {
	keys := new(keyList)

	for _, sec := range sections {
		s := keys.addSection(sec.Title)
		s.color = sec.Color

		for _, item := range sec.Items {
			if item.Key == "" {
				continue
			}

			s.addKey(item.Key, item.Desc)
		}
	}

	return formatKeymaps(keys, screenWidth)
}

func readerText(screenWidth int, withStoryNav bool) string {
	keys := new(keyList)

	nav := keys.addSection("Navigation")
	nav.addKey("j, k", "Down / up one line")
	nav.addKey("d, u", "Down / up half page")
	nav.addKey("␣, b", "Page down / up")
	nav.addKey("g, G", "Top / bottom")
	nav.addKey("n, N", "Next / prev section")
	nav.addKey("h, l", "Hide / show images")

	open := keys.addSection("Open")
	open.addKey("o", "Open story in browser")
	open.addKey("c", "Open comments in browser")

	if withStoryNav {
		open.addKey("J, K", "Open next / prev story")
	}

	app := keys.addSection("App")
	app.color = style.HeaderTertiary()
	app.addKey("i, ?", "Help")
	app.addKey("q, ⌫", "Back")

	return formatKeymaps(keys, screenWidth)
}

func commentText(screenWidth int, enableNerdFonts bool, withStoryNav bool) string {
	keys := new(keyList)

	read := keys.addSection("Read Mode")
	read.addKey("j, k", "Down / up one line")
	read.addKey("d, u", "Down / up half page")
	read.addKey("␣, b", "Page down / up")
	read.addKey("g, G", "Top / bottom")
	read.addKey("n, N", "Next / prev top comment")
	read.addKey("h, l", "Collapse / expand")
	read.addBreak()
	read.addKey("o", "Open story in browser")
	read.addKey("c", "Open comments in browser")

	if withStoryNav {
		read.addKey("J, K", "Open next / prev story")
	}

	read.addKey("↩", "Toggle all")
	read.addKey("⇥", "Navigate mode")
	read.addBreak()
	read.addKey("i, ?", "Help")
	read.addKey("q, ⌫", "Back")

	nav := keys.addSection("Navigate Mode")
	nav.addKey("j, k", "Next / prev comment")
	nav.addKey("d, u", "Down / up half page")
	nav.addKey("n, N", "Next / prev top comment")
	nav.addKey("h, l", "Collapse / expand")
	nav.addBreak()
	nav.addKey("o", "Open story in browser")
	nav.addKey("c", "Open comments in browser")

	if withStoryNav {
		nav.addKey("J, K", "Open next / prev story")
	}

	nav.addKey("↩", "Toggle collapse")
	nav.addKey("⇥", "Read mode")
	nav.addBreak()
	nav.addKey("i, ?", "Help")
	nav.addKey("q, ⌫", "Back")

	legend := keys.addSection("Legend")
	legend.addLabel(style.CommentOP(labelText("OP", enableNerdFonts)), "Original Poster")
	legend.addLabel(style.CommentGP(labelText("GP", enableNerdFonts)), "Grandparent Poster")
	legend.addLabel(style.CommentMod(labelText("mod", enableNerdFonts)), "Moderator")
	legend.addLabel(style.CommentNewIndicator("●"), "New comment indicator")

	return formatKeymaps(keys, screenWidth)
}

func formatKeymaps(keys *keyList, screenWidth int) string {
	contentWidth := min(panelOuterWidth, screenWidth-panelLeftMargin)
	body := keys.print(contentWidth)

	leftMargin := strings.Repeat(" ", panelLeftMargin)

	return style.PrefixLines(body, leftMargin)
}

func labelText(fallback string, enableNerdFonts bool) string {
	if enableNerdFonts {
		return nerdfonts.Author
	}

	return fallback
}
