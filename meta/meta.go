package meta

import (
	"fmt"
	"strconv"

	"clx/constants/nerdfonts"

	"clx/comment"
	"clx/constants/unicode"
	"clx/item"
	"clx/settings"
	"clx/syntax"

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
		getScore(c.Points, config.EnableNerdFonts)

	joined := lipgloss.JoinHorizontal(lipgloss.Left, leftColumn.Render(leftColumnText),
		rightColumn.Render(rightColumnText))

	return getHeadline(c.Title, config) + newParagraph + style.Render(url+joined+rootComment)
}

func getAuthor(author string, enableNerdFonts bool) string {
	if enableNerdFonts {
		authorLabel := fmt.Sprintf("%s %s", nerdfonts.Author, author)

		return Red(authorLabel).String()
	}

	return fmt.Sprintf("by %s", Red(author).String())
}

func getComments(commentsCount int, enableNerdFonts bool) string {
	comments := strconv.Itoa(commentsCount)

	if enableNerdFonts {
		commentsLabel := fmt.Sprintf("%s %s", nerdfonts.Comment, comments)

		return Magenta(commentsLabel).String()
	}

	return fmt.Sprintf("%s comments", Magenta(comments).String())
}

func getScore(points int, enableNerdFonts bool) string {
	score := strconv.Itoa(points)

	if enableNerdFonts {
		pointsLabel := fmt.Sprintf("%s %s", score, nerdfonts.Score)

		return Yellow(pointsLabel).String()
	}

	return fmt.Sprintf("%s points", Yellow(score).String())
}

func getID(id int, enableNerdFonts bool) string {
	if enableNerdFonts {
		return fmt.Sprintf("%d %s", Faint(id).Green(), Green(nerdfonts.Tag))
	}

	return fmt.Sprintf("%s %d", "ID", Faint(id).Green())
}

func getNewCommentsInfo(newComments int, enableNerdFonts bool) string {
	if newComments == 0 {
		return ""
	}

	comments := strconv.Itoa(newComments)

	if enableNerdFonts {
		return fmt.Sprintf(" (%s)", Cyan(comments).String())
	}

	return fmt.Sprintf(" (%s new)", Cyan(comments).String())
}

func getHeadline(title string, config *settings.Config) string {
	formattedTitle := highlightTitle(unicode.ZeroWidthSpace+" "+newLine+title, config.DisableHeadlineHighlighting,
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

	rootComment := comment.Print(c, config, config.CommentWidth-2, config.CommentWidth)
	wrappedComment, _ := text.Wrap(rootComment, config.CommentWidth-2)

	return newParagraph + wrappedComment
}
