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
func urlRow(url, domain string, contentWidth int) string {
	if domain == "" {
		return ""
	}

	display := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	display = strings.TrimSuffix(display, "/")

	return style.MetaURL(xansi.Truncate(display, contentWidth, "…"), url)
}

// byline is the opening rule's title. No author, no title — the frame falls
// back to a plain rule rather than heading the block with an empty byline.
func byline(author, timeAgo string, enableNerdFonts bool) string {
	if author == "" {
		return ""
	}

	return authorLabel(author, enableNerdFonts) + " " + style.Faint(timeAgo)
}

// statLabels is the group closing the opening rule: the comment tally, then
// the score against the right corner. The count sheds before the score when
// the rule runs out of room.
func statLabels(d Data) []string {
	return []string{
		commentsLabel(d.CommentsCount, d.NewComments, d.NerdFonts),
		scoreLabel(d.Points, d.NerdFonts),
	}
}

// scoreLabel leads with the arrow: a rule segment follows the score, so the
// label reads icon-then-value like the byline beside it.
func scoreLabel(points int, enableNerdFonts bool) string {
	score := strconv.Itoa(points)

	if enableNerdFonts {
		return style.MetaScore(fmt.Sprintf("%s %s", nerdfonts.Score, score))
	}

	return fmt.Sprintf("%s points", style.MetaScore(score))
}

// commentsLabel is the comment tally: total comments and, in parentheses,
// how many arrived since the last visit. The count takes the meta comments
// color the way the score takes its own, the new-comment count the meta
// new-comments color; the words and parentheses around them stay plain.
func commentsLabel(comments, newComments int, enableNerdFonts bool) string {
	label := style.MetaComments(strconv.Itoa(comments)) + " comments"
	if enableNerdFonts {
		label = style.MetaComments(fmt.Sprintf("%s %d", nerdfonts.Comment, comments))
	}

	switch {
	case newComments <= 0:
		return label
	case enableNerdFonts:
		return label + " (" + style.MetaNewComments(strconv.Itoa(newComments)) + ")"
	default:
		return label + " (" + style.MetaNewComments(strconv.Itoa(newComments)) + " new)"
	}
}

func authorLabel(author string, enableNerdFonts bool) string {
	if enableNerdFonts {
		return style.MetaAuthor(fmt.Sprintf("%s %s", nerdfonts.Author, author))
	}

	return fmt.Sprintf("by %s", style.MetaAuthor(author))
}
