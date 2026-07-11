package meta

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"

	xansi "github.com/charmbracelet/x/ansi"
)

// urlRow is the block's footer: the story link on the block's last row,
// truncated to a single ellipsis when it overruns the content width. The
// scheme is stripped from the display — the row is visibly a link already —
// but the hyperlink target keeps the full URL. Stories without a link
// (domain is empty) have no URL row at all.
func urlRow(url, domain string, contentWidth int, enableNerdFonts bool) string {
	if domain == "" {
		return ""
	}

	display := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")

	display = strings.TrimSuffix(display, "/")
	if enableNerdFonts {
		display = nerdfonts.Link + " " + display
	}

	return style.MetaURL(xansi.Truncate(display, contentWidth, "…"), url)
}

func byline(author, timeAgo string, enableNerdFonts bool) string {
	return authorLabel(author, enableNerdFonts) + " " + style.Faint(timeAgo)
}

// scoreLabel puts the arrow after the number: the score sits on the block's
// right edge, so the glyph belongs on the outside.
func scoreLabel(points int, enableNerdFonts bool) string {
	score := strconv.Itoa(points)

	if enableNerdFonts {
		return style.MetaScore(fmt.Sprintf("%s %s", score, nerdfonts.Score))
	}

	return fmt.Sprintf("%s points", style.MetaScore(score))
}

func authorLabel(author string, enableNerdFonts bool) string {
	if enableNerdFonts {
		return style.MetaAuthor(fmt.Sprintf("%s %s", nerdfonts.Author, author))
	}

	return fmt.Sprintf("by %s", style.MetaAuthor(author))
}
