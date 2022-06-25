package meta

import (
	"clx/constants/unicode"
	"clx/item"
	"clx/parser"
	"clx/settings"
	"clx/syntax"
	"strconv"

	text "github.com/MichaelMure/go-term-text"

	. "github.com/logrusorgru/aurora/v3"

	"github.com/charmbracelet/lipgloss"
)

const (
	newLine      = "\n"
	newParagraph = "\n\n"
)

func GetReaderModeMetaBlock(title string, url string, lineWidth int) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		PaddingRight(1).
		Width(lineWidth)

	formattedTitle, _ := text.Wrap(Bold(title).String(), lineWidth)
	formattedTitle = unicode.ZeroWidthSpace + newLine + formattedTitle
	formattedURL := Blue(text.TruncateMax(url, lineWidth-2)).String()
	info := newParagraph + Green("Reader Mode").String()

	return formattedTitle + newParagraph + style.Render(formattedURL+info) + newParagraph
}

func GetCommentSectionMetaBlock(c *item.Item, config *settings.Config, newComments int) string {
	columnWidth := config.CommentWidth/2 - 1
	url := getURL(c.URL, c.Domain, config.CommentWidth)
	rootComment := parseRootComment(c.Content, config)

	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		PaddingLeft(1).
		PaddingRight(1).
		Width(config.CommentWidth)

	leftColumn := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Left)
	leftColumnText := getAuthor(c.User, config.EnableNerdFonts) + " " + Faint(c.TimeAgo).String() + newLine +
		getComments(c.CommentsCount, config.EnableNerdFonts) + getNewCommentsInfo(newComments, config.EnableNerdFonts)

	rightColumn := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Right)
	rightColumnText := getID(c.ID, config.EnableNerdFonts) + newLine +
		getPoints(c.Points, config.EnableNerdFonts)

	joined := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn.Render(leftColumnText),
		rightColumn.Render(rightColumnText))

	return getHeadline(c.Title, config) + newParagraph + style.Render(url+joined+rootComment)
}

func getAuthor(author string, enableNerdFonts bool) string {
	if enableNerdFonts {
		return Red(" " + author).String()
	}

	return "by " + Red(author).String()
}

func getComments(commentsCount int, enableNerdFonts bool) string {
	comments := strconv.Itoa(commentsCount)

	if enableNerdFonts {
		return Magenta(" " + comments).String()
	}

	return Magenta(comments).String() + " comments"
}

func getPoints(points int, enableNerdFonts bool) string {
	p := strconv.Itoa(points)

	if enableNerdFonts {
		return Yellow(p + " ﰵ").String()
	}

	return Yellow(p).String() + " points"
}

func getID(id int, enableNerdFonts bool) string {
	idTag := strconv.Itoa(id)

	if enableNerdFonts {
		return Green(idTag + " ").Faint().String()
	}

	return Green("ID " + idTag).Faint().String()
}

func getNewCommentsInfo(newComments int, enableNerdFonts bool) string {
	if newComments == 0 {
		return ""
	}

	c := strconv.Itoa(newComments)

	if enableNerdFonts {
		return " (" + Cyan(c).String() + ")"
	}

	return " (" + Cyan(c).String() + " new)"
}

func getHeadline(title string, config *settings.Config) string {
	formattedTitle := highlightTitle(unicode.ZeroWidthSpace+" "+newLine+title, config.HighlightHeadlines,
		config.EnableNerdFonts)
	wrappedHeadline, _ := text.Wrap(formattedTitle, config.CommentWidth)

	return wrappedHeadline
}

func highlightTitle(title string, highlightHeadlines bool, enableNerdFont bool) string {
	highlightedTitle := title

	if highlightHeadlines {
		highlightedTitle = syntax.HighlightYCStartupsInHeadlines(highlightedTitle, syntax.HeadlineInCommentSection, enableNerdFont)
		highlightedTitle = syntax.HighlightYear(highlightedTitle, syntax.HeadlineInCommentSection, enableNerdFont)
		highlightedTitle = syntax.HighlightHackerNewsHeadlines(highlightedTitle, syntax.HeadlineInCommentSection)
		highlightedTitle = syntax.HighlightSpecialContent(highlightedTitle, syntax.HeadlineInCommentSection, enableNerdFont)
	}

	return Bold(highlightedTitle).String()
}

func getURL(url string, domain string, lineWidth int) string {
	if domain == "" {
		return ""
	}

	truncatedURL := text.TruncateMax(url, lineWidth-2)
	formattedURL := Blue(truncatedURL).String() + newLine

	return formattedURL + newLine
}

func parseRootComment(c string, config *settings.Config) string {
	if c == "" {
		return ""
	}

	comment := parser.ParseComment(c, config, config.CommentWidth-2, config.CommentWidth)
	wrappedComment, _ := text.Wrap(comment, config.CommentWidth-2)

	return newParagraph + wrappedComment
}
