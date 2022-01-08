package meta

import (
	"clx/constants/unicode"
	"clx/core"
	"clx/item"
	"clx/parser"
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
	formattedUrl := Blue(text.TruncateMax(url, lineWidth-2)).String()
	info := newParagraph + Green("Reader Mode").String()

	// formattedTitle := Bold(title).String()
	// formattedUrl := getURL(url, "_", lineWidth)

	return formattedTitle + newParagraph + style.Render(formattedUrl+info) + newParagraph
}

func GetCommentSectionMetaBlock(c *item.Item, config *core.Config) string {
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
		Yellow(points).String() + " points"

	rightColumn := lipgloss.NewStyle().
		Width(columnWidth).
		Align(lipgloss.Right)
	rightColumnText := Green("ID "+id).Faint().String() + newLine +
		Magenta(numberOfComments).String() + " comments"

	joined := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn.Render(leftColumnText),
		rightColumn.Render(rightColumnText))

	return getHeadline(c.Title, config) + newParagraph + style.Render(url+joined+rootComment)
}

func getHeadline(title string, config *core.Config) string {
	formattedTitle := highlightTitle(unicode.ZeroWidthSpace+" "+newLine+title, config.HighlightHeadlines)
	wrappedHeadline, _ := text.Wrap(formattedTitle, config.CommentWidth)

	// wrappedHeadline := wordwrap.String(formattedTitle, config.CommentWidth)

	return wrappedHeadline
}

func highlightTitle(title string, highlightHeadlines bool) string {
	highlightedTitle := title

	if highlightHeadlines {
		highlightedTitle = syntax.HighlightYCStartupsInHeadlines(highlightedTitle)
		highlightedTitle = syntax.HighlightYearInHeadlines(highlightedTitle)
		highlightedTitle = syntax.HighlightHackerNewsHeadlines(highlightedTitle)
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

func parseRootComment(c string, config *core.Config) string {
	if c == "" {
		return ""
	}

	comment := parser.ParseComment(c, config, config.CommentWidth-2, config.CommentWidth)
	wrappedComment, _ := text.Wrap(comment, config.CommentWidth-2)

	return newParagraph + wrappedComment
}
