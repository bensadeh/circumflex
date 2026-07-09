package meta

import (
	"fmt"
	"strconv"

	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

// urlLine is the block's header: the story link over a blank separator line.
// Stories without a link (domain is empty) have no URL rows at all.
func urlLine(url, domain string, contentWidth int) string {
	if domain == "" {
		return ""
	}

	truncatedURL := xansi.Truncate(url, contentWidth, "")

	return style.MetaURL(truncatedURL, url) + "\n\n"
}

func byline(author, timeAgo string, enableNerdFonts bool) string {
	return authorLabel(author, enableNerdFonts) + " " + style.Faint(timeAgo)
}

func authorLabel(author string, enableNerdFonts bool) string {
	if enableNerdFonts {
		return style.MetaAuthor(fmt.Sprintf("%s %s", nerdfonts.Author, author))
	}

	return fmt.Sprintf("by %s", style.MetaAuthor(author))
}

func readerModeLabel(enableNerdFonts bool) string {
	if enableNerdFonts {
		return style.MetaReaderMode(nerdfonts.Document + " Reader Mode")
	}

	return style.MetaReaderMode("Reader Mode")
}

func commentsLabel(commentsCount int, enableNerdFonts bool) string {
	comments := strconv.Itoa(commentsCount)

	if enableNerdFonts {
		return style.MetaComments(fmt.Sprintf("%s %s", nerdfonts.Comment, comments))
	}

	return fmt.Sprintf("%s comments", style.MetaComments(comments))
}

func newCommentsLabel(newComments int, enableNerdFonts bool) string {
	if newComments == 0 {
		return ""
	}

	comments := strconv.Itoa(newComments)

	if enableNerdFonts {
		return fmt.Sprintf(" (%s)", style.MetaNewComments(comments))
	}

	return fmt.Sprintf(" (%s new)", style.MetaNewComments(comments))
}

func scoreLabel(points int, enableNerdFonts bool) string {
	score := strconv.Itoa(points)

	if enableNerdFonts {
		return style.MetaScore(fmt.Sprintf("%s %s", score, nerdfonts.Score))
	}

	return fmt.Sprintf("%s points", style.MetaScore(score))
}

func idLabel(id int, enableNerdFonts bool) string {
	idStr := lipgloss.NewStyle().Faint(true).Foreground(style.MetaIDColor()).Render(strconv.Itoa(id))

	if enableNerdFonts {
		return fmt.Sprintf("%s %s", idStr, lipgloss.NewStyle().Foreground(style.MetaIDColor()).Render(nerdfonts.Tag))
	}

	return fmt.Sprintf("%s %s", "ID", idStr)
}
