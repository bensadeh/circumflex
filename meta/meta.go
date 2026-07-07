package meta

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
)

const (
	newLine      = "\n"
	newParagraph = "\n\n"

	borderSize  = 2
	paddingSize = 2
	boxOverhead = borderSize + paddingSize
)

func ReaderModeMetaBlock(url, author, timeAgo string, id, points int, enableNerdFonts bool, width int) string {
	bottomLeft := readerModeLabel(enableNerdFonts)

	return metaBlock(url, url, author, timeAgo, id, points, enableNerdFonts, bottomLeft, "", width) + newParagraph
}

func ReaderModeURLBlock(url string, enableNerdFonts bool, width int) string {
	s := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		PaddingRight(1).
		MarginLeft(1).
		Width(width + borderSize)

	return s.Render(urlLine(url, url, width-paddingSize)+readerModeLabel(enableNerdFonts)) + newParagraph
}

func CommentSectionMetaBlock(url, domain, author, timeAgo string, id, commentsCount, points, newComments int, enableNerdFonts bool, rootComment string, width int) string {
	bottomLeft := commentsLabel(commentsCount, enableNerdFonts) + newCommentsLabel(newComments, enableNerdFonts)

	return metaBlock(url, domain, author, timeAgo, id, points, enableNerdFonts, bottomLeft, rootComment, width)
}

// PlaceholderMetaBlock is the meta block's loading stand-in: an empty, dimmed
// box with the same dimensions as the block the loaded view will draw, so the
// box sits in place from the first frame and the content fills it in without
// moving it. hasURL mirrors urlLine: stories without a link have no URL rows
// to reserve.
func PlaceholderMetaBlock(width int, hasURL bool) string {
	s := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		PaddingRight(1).
		MarginLeft(1).
		Width(width + borderSize)

	contentRows := 2
	if hasURL {
		contentRows = 4
	}

	lines := strings.Split(s.Render(strings.Repeat(newLine, contentRows-1)), newLine)
	for i, line := range lines {
		lines[i] = style.Faint(line)
	}

	return strings.Join(lines, newLine)
}

func metaBlock(url, domain, author, timeAgo string, id, points int, enableNerdFonts bool, bottomLeft, footer string, width int) string {
	s := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		PaddingRight(1).
		MarginLeft(1).
		Width(width + borderSize)

	contentWidth := width - paddingSize
	columnWidth := contentWidth / 2

	header := urlLine(url, domain, contentWidth)

	leftColumn := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Left)
	leftColumnText := authorLabel(author, enableNerdFonts) + " " + style.Faint(timeAgo) + newLine +
		bottomLeft

	rightColumn := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Right)
	rightColumnText := idLabel(id, enableNerdFonts) + newLine +
		scoreLabel(points, enableNerdFonts)

	joined := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn.Render(leftColumnText),
		rightColumn.Render(rightColumnText))

	return s.Render(header + joined + footer)
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
		commentsLabel := fmt.Sprintf("%s %s", nerdfonts.Comment, comments)

		return style.MetaComments(commentsLabel)
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

func authorLabel(author string, enableNerdFonts bool) string {
	if enableNerdFonts {
		authorLabel := fmt.Sprintf("%s %s", nerdfonts.Author, author)

		return style.MetaAuthor(authorLabel)
	}

	return fmt.Sprintf("by %s", style.MetaAuthor(author))
}

func scoreLabel(points int, enableNerdFonts bool) string {
	score := strconv.Itoa(points)

	if enableNerdFonts {
		pointsLabel := fmt.Sprintf("%s %s", score, nerdfonts.Score)

		return style.MetaScore(pointsLabel)
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

func urlLine(url, domain string, contentWidth int) string {
	if domain == "" {
		return ""
	}

	truncatedURL := xansi.Truncate(url, contentWidth, "")
	formattedURL := style.MetaURL(truncatedURL, url) + newLine

	return formattedURL + newLine
}
