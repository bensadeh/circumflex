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

	points := strconv.Itoa(c.Points)
	numberOfComments := strconv.Itoa(c.CommentsCount)
	id := strconv.Itoa(c.ID)

	leftColumn := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Left)
	leftColumnText := "by " + Red(c.User).String() + " " + Faint(c.TimeAgo).String() + newLine +
		Magenta(numberOfComments).String() + " comments" + getNewCommentsInfo(newComments)

	rightColumn := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Right)
	rightColumnText := Green("ID "+id).Faint().String() + newLine +
		Yellow(points).String() + " points"

	joined := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn.Render(leftColumnText),
		rightColumn.Render(rightColumnText))

	return getHeadline(c.Title, config) + newParagraph + style.Render(url+joined+rootComment)
}

func getNewCommentsInfo(newComments int) string {
	if newComments == 0 {
		return ""
	}

	c := strconv.Itoa(newComments)

	return " (" + Cyan(c).String() + " new)"
}

func getHeadline(title string, config *settings.Config) string {
	formattedTitle := highlightTitle(unicode.ZeroWidthSpace+" "+newLine+title, config.HighlightHeadlines)
	wrappedHeadline, _ := text.Wrap(formattedTitle, config.CommentWidth)

	return wrappedHeadline
}

func highlightTitle(title string, highlightHeadlines bool) string {
	highlightedTitle := title

	if highlightHeadlines {
		highlightedTitle = syntax.HighlightYCStartupsInHeadlines(highlightedTitle, syntax.Bold)
		highlightedTitle = syntax.HighlightYearInHeadlines(highlightedTitle, syntax.Bold)
		highlightedTitle = syntax.HighlightHackerNewsHeadlines(highlightedTitle, syntax.Bold)
		highlightedTitle = syntax.HighlightSpecialContent(highlightedTitle)
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
