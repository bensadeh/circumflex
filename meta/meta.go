package meta

import (
	"clx/comment"
	"clx/constants"
	"clx/item"
	"clx/nerdfonts"
	"clx/settings"
	"clx/style"
	"clx/syntax"
	"fmt"
	"strconv"

	text "github.com/MichaelMure/go-term-text"

	"charm.land/lipgloss/v2"
)

const (
	newLine      = "\n"
	newParagraph = "\n\n"
)

func GetReaderModeMetaBlock(title string, url string, lineWidth int) string {
	s := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		PaddingRight(1).
		Width(lineWidth)

	contentWidth := lineWidth - s.GetHorizontalBorderSize() - s.GetHorizontalPadding()

	formattedTitle, _ := text.Wrap(style.Bold(title), lineWidth)
	formattedTitle = constants.InvisibleCharacterForTopLevelComments + newLine + formattedTitle
	formattedURL := style.MetaURL(text.TruncateMax(url, contentWidth))
	info := newParagraph + style.MetaReaderMode("Reader Mode")

	return formattedTitle + newParagraph + s.Render(formattedURL+info) + newParagraph
}

func GetCommentSectionMetaBlock(c *item.Story, config *settings.Config, newComments int) string {
	s := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		PaddingRight(1).
		Width(config.CommentWidth)

	contentWidth := config.CommentWidth - s.GetHorizontalBorderSize() - s.GetHorizontalPadding()
	columnWidth := contentWidth / 2

	url := getURL(c.URL, c.Domain, contentWidth)
	rootComment := parseRootComment(c.Content, config, contentWidth)

	leftColumn := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Left)
	leftColumnText := getAuthor(c.User, config.EnableNerdFonts) + " " + style.Faint(c.TimeAgo) + newLine +
		getComments(c.CommentsCount, config.EnableNerdFonts) + getNewCommentsInfo(newComments, config.EnableNerdFonts)

	rightColumn := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Right)
	rightColumnText := getID(c.ID, config.EnableNerdFonts) + newLine +
		getScore(c.Points, config.EnableNerdFonts)

	joined := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn.Render(leftColumnText),
		rightColumn.Render(rightColumnText))

	return getHeadline(c.Title, config) + newParagraph + s.Render(url+joined+rootComment)
}

func getAuthor(author string, enableNerdFonts bool) string {
	if enableNerdFonts {
		authorLabel := fmt.Sprintf("%s %s", nerdfonts.Author, author)

		return style.MetaAuthor(authorLabel)
	}

	return fmt.Sprintf("by %s", style.MetaAuthor(author))
}

func getComments(commentsCount int, enableNerdFonts bool) string {
	comments := strconv.Itoa(commentsCount)

	if enableNerdFonts {
		commentsLabel := fmt.Sprintf("%s %s", nerdfonts.Comment, comments)

		return style.MetaComments(commentsLabel)
	}

	return fmt.Sprintf("%s comments", style.MetaComments(comments))
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

func getNewCommentsInfo(newComments int, enableNerdFonts bool) string {
	if newComments == 0 {
		return ""
	}

	comments := strconv.Itoa(newComments)

	if enableNerdFonts {
		return fmt.Sprintf(" (%s)", style.MetaNewComments(comments))
	}

	return fmt.Sprintf(" (%s new)", style.MetaNewComments(comments))
}

func getHeadline(title string, config *settings.Config) string {
	formattedTitle := highlightTitle(constants.InvisibleCharacterForTopLevelComments+" "+newLine+title, config.DisableHeadlineHighlighting,
		config.EnableNerdFonts)
	wrappedHeadline, _ := text.Wrap(formattedTitle, config.CommentWidth)

	return wrappedHeadline
}

func highlightTitle(title string, disableHeadlineHighlighting bool, enableNerdFont bool) string {
	highlightedTitle := title

	if !disableHeadlineHighlighting {
		highlightedTitle = syntax.HighlightYCStartupsInHeadlines(highlightedTitle, syntax.HeadlineInCommentSection, enableNerdFont)
		highlightedTitle = syntax.HighlightYear(highlightedTitle, syntax.HeadlineInCommentSection)
		highlightedTitle = syntax.HighlightHackerNewsHeadlines(highlightedTitle, syntax.HeadlineInCommentSection)
		highlightedTitle = syntax.HighlightSpecialContent(highlightedTitle, syntax.HeadlineInCommentSection, enableNerdFont)
	}

	return style.Bold(highlightedTitle)
}

func getURL(url string, domain string, contentWidth int) string {
	if domain == "" {
		return ""
	}

	truncatedURL := text.TruncateMax(url, contentWidth)
	formattedURL := style.MetaURL(truncatedURL) + newLine

	return formattedURL + newLine
}

func parseRootComment(c string, config *settings.Config, contentWidth int) string {
	if c == "" {
		return ""
	}

	rootComment := comment.Print(c, config, contentWidth, contentWidth)
	wrappedComment, _ := text.Wrap(rootComment, contentWidth)

	return newParagraph + wrappedComment
}
