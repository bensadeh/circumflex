package tree

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"clx/constants/nerdfonts"

	"clx/comment"
	"clx/constants/margins"
	"clx/constants/unicode"
	"clx/item"
	"clx/meta"
	"clx/settings"
	"clx/syntax"
	"clx/tree/postprocessor"

	. "github.com/logrusorgru/aurora/v3"

	text "github.com/MichaelMure/go-term-text"
)

const (
	newLine      = "\n"
	newParagraph = "\n\n"
)

func Print(comments *item.Item, config *settings.Config, screenWidth int, lastVisited int64) string {
	commentSectionScreenWidth := screenWidth - margins.CommentSectionLeftMargin

	header := getHeader(comments, config, lastVisited)
	firstCommentID := getFirstCommentID(comments.Comments)

	replies := ""

	for _, reply := range comments.Comments {
		replies += printReplies(reply, config, commentSectionScreenWidth, comments.User, "", firstCommentID,
			lastVisited)
	}

	commentSection := postprocessor.Process(header+replies+newLine, screenWidth)

	return commentSection
}

func getFirstCommentID(comments []*item.Item) int {
	if len(comments) == 0 {
		return 0
	}

	return comments[0].ID
}

func getHeader(c *item.Item, config *settings.Config, lastVisited int64) string {
	newComments := getNewCommentsCount(c, lastVisited)

	return meta.GetCommentSectionMetaBlock(c, config, newComments) + newParagraph
}

func printReplies(c *item.Item, config *settings.Config, screenWidth int, originalPoster string,
	parentPoster string, firstCommentID int, lastVisited int64,
) string {
	isDeletedAndHasNoReplies := c.Content == "[deleted]" && len(c.Comments) == 0
	if isDeletedAndHasNoReplies {
		return ""
	}

	indentation := getIndentString(c.Level)
	indentSize := len(indentation)
	availableScreenWidth := screenWidth - indentSize - margins.CommentSectionLeftMargin
	adjustedCommentWidth := config.CommentWidth - c.Level

	comment := formatComment(c, config, originalPoster, parentPoster, adjustedCommentWidth, availableScreenWidth,
		lastVisited)
	indentedComment, _ := text.WrapWithPad(comment, screenWidth, indentation)
	fullComment := getSeparator(c.Level, config.CommentWidth, c.ID, firstCommentID) + indentedComment + newLine
	fullComment += getButton(c.Level, getReplyCount(c), config.CommentWidth)

	fullCommentWithFilterTag := addFilterTag(c.Level, fullComment)

	if c.Level == 0 {
		parentPoster = c.User
	}

	for _, reply := range c.Comments {
		fullCommentWithFilterTag += printReplies(reply, config, screenWidth, originalPoster, parentPoster, firstCommentID,
			lastVisited)
	}

	return fullCommentWithFilterTag
}

func getButton(level int, replyCount int, commentWidth int) string {
	if replyCount == 0 || level != 0 {
		return ""
	}

	replies := ""

	if replyCount == 1 {
		replies = "reply"
	} else {
		replies = "replies"
	}

	buttonLabel := fmt.Sprintf("%d %s", replyCount, replies)

	buttonNotPressedStyle := lipgloss.NewStyle().
		Bold(true).
		AlignHorizontal(lipgloss.Center).
		SetString("▶ " + buttonLabel)

	buttonNotPressed := buttonNotPressedStyle.String()

	buttonPressedStyle := lipgloss.NewStyle().
		Bold(true).
		Faint(true).
		AlignHorizontal(lipgloss.Center).
		SetString("▼ " + buttonLabel)

	buttonPressed := buttonPressedStyle.String()

	style := lipgloss.NewStyle().Width(commentWidth).AlignHorizontal(lipgloss.Center)

	return newLine + style.Render(buttonNotPressed) + unicode.AnotherInvisibleCharacter +
		newLine + style.Render(buttonPressed) + unicode.InvisibleCharacter + newLine
}

func addFilterTag(level int, fullComment string) string {
	if level == 0 {
		return fullComment
	}

	fullCommentWithFilterTags := ""
	lines := strings.Split(fullComment, "\n")

	for i, line := range lines {
		if i == len(lines)-1 {
			continue
		}
		fullCommentWithFilterTags += line + unicode.InvisibleCharacter + "\n"
	}

	return fullCommentWithFilterTags
}

func formatComment(c *item.Item, config *settings.Config, originalPoster string, parentPoster string, commentWidth int,
	availableScreenWidth int, lastVisited int64,
) string {
	coloredIndentSymbol := syntax.ColorizeIndentSymbol(config.IndentationSymbol, c.Level)

	header := getCommentHeader(c, originalPoster, parentPoster, lastVisited, config)
	formattedComment := comment.Print(c.Content, config, commentWidth, availableScreenWidth)

	paddedComment, _ := text.WrapWithPad(formattedComment, availableScreenWidth, coloredIndentSymbol)

	return header + paddedComment
}

func getSeparator(level int, commentWidth int, currentCommentID int, firstCommentID int) string {
	if currentCommentID == firstCommentID {
		return ""
	}

	if level != 0 || currentCommentID == firstCommentID {
		return newLine
	}

	return Faint(strings.Repeat("▁", commentWidth)).Faint().String() + newLine + newLine
}

func getIndentString(level int) string {
	if level == 0 {
		return ""
	}

	return strings.Repeat(" ", level-1)
}

func getCommentHeader(c *item.Item, originalPoster string, parentPoster string, lastVisited int64, config *settings.Config) string {
	if c.Level == 0 {
		return formatHeader(c, originalPoster, parentPoster, true,
			0, lastVisited, config)
	}

	return formatHeader(c, originalPoster, parentPoster, false,
		1, lastVisited, config)
}

func formatHeader(c *item.Item, originalPoster string, parentPoster string,
	enableZeroWidthSpace bool, indentSize int, lastVisited int64, config *settings.Config,
) string {
	author := getAuthor(c.User, lastVisited, c.Time)
	authorLabel := getAuthorLabel(c.User, originalPoster, parentPoster, config.EnableNerdFonts)
	zeroWidthSpace := getZeroWidthSpace(enableZeroWidthSpace)
	indentation := strings.Repeat(" ", indentSize)

	return zeroWidthSpace + indentation + author + authorLabel +
		Faint(c.TimeAgo).String() + newLine
}

func getAuthor(author string, lastVisited, timePosted int64) string {
	authorInBold := Bold(author).String() + " "

	commentIsNew := lastVisited < timePosted

	if commentIsNew {
		return authorInBold + Cyan("●").String() + " "
	}

	return authorInBold
}

func getZeroWidthSpace(enabled bool) string {
	if enabled {
		return unicode.ZeroWidthSpace
	}

	return ""
}

func getReplyCount(comments *item.Item) int {
	numberOfReplies := 0

	return incrementReplyCount(comments, &numberOfReplies)
}

func incrementReplyCount(comments *item.Item, repliesSoFar *int) int {
	for _, reply := range comments.Comments {
		*repliesSoFar++
		incrementReplyCount(reply, repliesSoFar)
	}

	return *repliesSoFar
}

func getNewCommentsCount(comments *item.Item, lastVisited int64) int {
	numberOfReplies := 0

	return incrementNewCommentsCount(comments, &numberOfReplies, lastVisited)
}

func incrementNewCommentsCount(comments *item.Item, newCommentsSoFar *int, lastVisited int64) int {
	for _, reply := range comments.Comments {
		commentIsNew := lastVisited < reply.Time
		if commentIsNew {
			*newCommentsSoFar++
		}

		incrementNewCommentsCount(reply, newCommentsSoFar, lastVisited)
	}

	return *newCommentsSoFar
}

func getAuthorLabel(author, originalPoster, parentPoster string, enableNerdFonts bool) string {
	if enableNerdFonts {
		authorLabel := nerdfonts.Author + " "

		switch author {
		case "dang":
			return Green(authorLabel).String()

		case originalPoster:
			return Red(authorLabel).String()

		case parentPoster:
			return Magenta(authorLabel).String()

		default:
			return ""
		}
	}

	switch author {
	case "dang":
		return Green("mod ").String()

	case originalPoster:
		return Red("OP ").String()

	case parentPoster:
		return Magenta("PP ").String()

	default:
		return ""
	}
}
