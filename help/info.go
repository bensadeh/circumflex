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

// MainMenuPanelRightEdge is the column where the front-page help panels'
// right border sits; the search header right-aligns its date group to it so
// the two screens share an edge.
func MainMenuPanelRightEdge(screenWidth int) int {
	return panelLeftMargin + mainMenuContentWidth(screenWidth)
}

func mainMenuText(screenWidth int, sections []Section) string {
	keys := new(keyList)

	for _, sec := range sections {
		s := keys.addSection(sec.Title)

		for _, item := range sec.Items {
			if item.Key == "" {
				continue
			}

			s.addKey(item.Key, item.Desc)
		}
	}

	return formatKeymaps(keys, panelLeftMargin, mainMenuContentWidth(screenWidth))
}

func readerText(leftMargin, contentWidth int, inApp bool) string {
	keys := new(keyList)

	nav := keys.addSection("Navigation")
	nav.addKey("j, k", "Down / up one line")
	nav.addKey("d, u", "Down / up half page")
	nav.addKey("␣, b", "Page down / up")
	nav.addKey("g, G", "Top / bottom")
	nav.addKey("n, N", "Next / prev section")
	nav.addKey("h, l", "Hide / show images")

	links := keys.addSection("Link Selector")
	links.addKey("⇥", "Enter / exit selector")
	links.addKey("j/n, k/N", "Next / prev link")
	links.addKey("↩", "Open link in place")

	search := keys.addSection("Search")
	search.addKey("/", "Search article")
	search.addKey("n, N", "Next / prev match")
	search.addKey("esc", "Clear search")

	app := keys.addSection("App")
	app.addKey("o", "Open story in browser")
	app.addKey("c", "Open comments in browser")

	if inApp {
		app.addKey("J, K", "Open next / prev story")
	}

	app.addBreak()

	if inApp {
		app.addKey("z", "Toggle wide layout")
	}

	app.addKey("i, ?", "Help")
	app.addKey("q, ⌫", "Back")

	return formatKeymaps(keys, leftMargin, contentWidth)
}

func commentText(leftMargin, contentWidth int, enableNerdFonts bool, inApp bool) string {
	keys := new(keyList)

	read := keys.addSection("Read Mode")
	read.addKey("j, k", "Down / up one line")
	read.addKey("d, u", "Down / up half page")
	read.addKey("␣, b", "Page down / up")
	read.addKey("g, G", "Top / bottom")
	read.addKey("n, N", "Next / prev top comment")
	read.addKey("h, l", "Collapse / expand")
	read.addBreak()
	read.addKey("↩", "Toggle all")
	read.addKey("⇥", "Navigate mode")

	nav := keys.addSection("Navigate Mode")
	nav.addKey("j, k", "Next / prev comment")
	nav.addKey("d, u", "Down / up half page")
	nav.addKey("n, N", "Next / prev top comment")
	nav.addKey("h, l", "Collapse / expand")
	nav.addBreak()
	nav.addKey("↩", "Toggle collapse")
	nav.addKey("⇥", "Read mode")

	search := keys.addSection("Search")
	search.addKey("/", "Search all comments")
	search.addKey("n, N", "Next / prev match")
	search.addKey("esc", "Clear search")

	app := keys.addSection("App")
	app.addKey("o", "Open story in browser")
	app.addKey("c", "Open comments in browser")

	if inApp {
		app.addKey("J, K", "Open next / prev story")
	}

	app.addBreak()

	if inApp {
		app.addKey("z", "Toggle wide layout")
	}

	app.addKey("i, ?", "Help")
	app.addKey("q, ⌫", "Back")

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
