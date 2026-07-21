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
// label reads icon-then-value like the byline beside it. The color sits on
// the glyph when there is one and on the number only when the word has to
// carry the meaning alone.
func scoreLabel(points int, enableNerdFonts bool) string {
	score := strconv.Itoa(points)

	if enableNerdFonts {
		return style.MetaScore(nerdfonts.Score) + " " + score
	}

	return fmt.Sprintf("%s points", style.MetaScore(score))
}

// commentsLabel is the comment tally: total comments and, in parentheses,
// how many arrived since the last visit. The meta comments color sits on the
// glyph, or on the number when the word has to carry the meaning alone; the
// new-comment count takes the meta new-comments color either way, with the
// parentheses around it plain.
func commentsLabel(comments, newComments int, enableNerdFonts bool) string {
	label := style.MetaComments(strconv.Itoa(comments)) + " comments"
	if enableNerdFonts {
		label = style.MetaComments(nerdfonts.Comment) + " " + strconv.Itoa(comments)
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

// idLabel is the story's item id, the group closing the bottom rule against
// its right corner. The color sits on the glyph when there is one and on the
// number when the word "ID" has to carry the meaning alone, the same way the
// score and comment labels above it color the icon or the value. Reads
// "ID 12345" either way. A story with no id (id is zero) has no label and
// leaves the closing rule plain.
func idLabel(id int, enableNerdFonts bool) string {
	if id <= 0 {
		return ""
	}

	if enableNerdFonts {
		return style.MetaID(nerdfonts.ID) + " " + strconv.Itoa(id)
	}

	return "ID " + style.MetaID(strconv.Itoa(id))
}

func authorLabel(author string, enableNerdFonts bool) string {
	if enableNerdFonts {
		return style.MetaAuthor(fmt.Sprintf("%s %s", nerdfonts.Author, author))
	}

	return fmt.Sprintf("by %s", style.MetaAuthor(author))
}
