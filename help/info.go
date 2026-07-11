package help

import (
	"strings"

	"github.com/bensadeh/circumflex/layout"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"
)

// Main-menu help geometry: the same left margin as the comment section and
// reader views, and the default comment column's width, so the main help
// sits exactly where the detail views' help screens sit. The reader and
// comment help screens inherit their view's live geometry instead.
const (
	panelOuterWidth = 70
	panelLeftMargin = layout.CommentSectionLeftMargin
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

	return formatKeymaps(keys, panelLeftMargin, mainMenuContentWidth(screenWidth))
}

func readerText(leftMargin, contentWidth int, withStoryNav bool) string {
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

	return formatKeymaps(keys, leftMargin, contentWidth)
}

func commentText(leftMargin, contentWidth int, enableNerdFonts bool, withStoryNav bool) string {
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

	return formatKeymaps(keys, leftMargin, contentWidth)
}

func mainMenuContentWidth(screenWidth int) int {
	return min(panelOuterWidth, screenWidth-panelLeftMargin)
}

func formatKeymaps(keys *keyList, leftMargin, contentWidth int) string {
	body := keys.print(contentWidth)

	return style.PrefixLines(body, strings.Repeat(" ", leftMargin))
}

func labelText(fallback string, enableNerdFonts bool) string {
	if enableNerdFonts {
		return nerdfonts.Author
	}

	return fallback
}
