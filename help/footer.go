package help

import (
	"strings"

	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"
	"github.com/bensadeh/circumflex/version"
	"github.com/bensadeh/circumflex/view/pane"

	xansi "github.com/charmbracelet/x/ansi"
)

// Footer takes the geometry of the view it renders under, so its icons sit in
// the same columns as the view's own footer icons: the GitHub icon on the
// left margin, the version's tag icon ending at the column's right edge.
func Footer(leftMargin, contentWidth int, enableNerdFonts bool) string {
	if contentWidth <= 0 {
		return ""
	}

	url := style.Faint("github.com/bensadeh/circumflex")
	ver := style.Faint("version " + version.Version)

	if enableNerdFonts {
		// Nerd font glyphs render wider than one cell, so they get extra room.
		url = nerdfonts.GitHub + "  " + url
		ver = style.Faint(version.Version) + " " + nerdfonts.Tag
	}

	line := pane.FooterSections(contentWidth, url, ver)

	return strings.Repeat(" ", leftMargin) + xansi.Truncate(line, contentWidth, "")
}

// MainMenuFooter follows the main-menu help panels' own geometry.
func MainMenuFooter(screenWidth int, enableNerdFonts bool) string {
	return Footer(panelLeftMargin, mainMenuContentWidth(screenWidth), enableNerdFonts)
}
