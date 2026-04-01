package meta

import (
	"clx/nerdfonts"
	"clx/style"
	"fmt"
	"strconv"

	text "github.com/MichaelMure/go-term-text"

	"charm.land/lipgloss/v2"
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

func CommentSectionMetaBlock(url, domain, author, timeAgo string, id, commentsCount, points, newComments int, enableNerdFonts bool, rootComment string, width int) string {
	bottomLeft := commentsLabel(commentsCount, enableNerdFonts) + newCommentsLabel(newComments, enableNerdFonts)

	return metaBlock(url, domain, author, timeAgo, id, points, enableNerdFonts, bottomLeft, rootComment, width)
}

func metaBlock(url, domain, author, timeAgo string, id, points int, enableNerdFonts bool, bottomLeft, footer string, width int) string {
	s := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		PaddingRight(1).
		Width(width)

	contentWidth := width - boxOverhead
	columnWidth := contentWidth / 2

	urlLine := getURL(url, domain, contentWidth)

	leftColumn := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Left)
	leftColumnText := getAuthor(author, enableNerdFonts) + " " + style.Faint(timeAgo) + newLine +
		bottomLeft

	rightColumn := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Right)
	rightColumnText := getID(id, enableNerdFonts) + newLine +
		getScore(points, enableNerdFonts)

	joined := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn.Render(leftColumnText),
		rightColumn.Render(rightColumnText))

	return s.Render(urlLine + joined + footer)
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

func getAuthor(author string, enableNerdFonts bool) string {
	if enableNerdFonts {
		authorLabel := fmt.Sprintf("%s %s", nerdfonts.Author, author)

		return style.MetaAuthor(authorLabel)
	}

	return fmt.Sprintf("by %s", style.MetaAuthor(author))
}

func getScore(points int, enableNerdFonts bool) string {
	score := strconv.Itoa(points)

	if enableNerdFonts {
		pointsLabel := fmt.Sprintf("%s %s", score, nerdfonts.Score)

		return style.MetaScore(pointsLabel)
	}

	return fmt.Sprintf("%s points", style.MetaScore(score))
}

func getID(id int, enableNerdFonts bool) string {
	idStr := lipgloss.NewStyle().Faint(true).Foreground(style.MetaIDColor()).Render(strconv.Itoa(id))

	if enableNerdFonts {
		return fmt.Sprintf("%s %s", idStr, lipgloss.NewStyle().Foreground(style.MetaIDColor()).Render(nerdfonts.Tag))
	}

	return fmt.Sprintf("%s %s", "ID", idStr)
}

func getURL(url, domain string, contentWidth int) string {
	if domain == "" {
		return ""
	}

	truncatedURL := text.TruncateMax(url, contentWidth)
	formattedURL := style.MetaURL(truncatedURL) + newLine

	return formattedURL + newLine
}
