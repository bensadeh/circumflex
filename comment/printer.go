package comment

import (
	"clx/constants/messages"
	"clx/constants/unicode"
	"clx/core"
	"clx/endpoints"
	"clx/indent"
	"clx/syntax"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora/v3"

	text "github.com/MichaelMure/go-term-text"
)

const (
	newLine      = "\n"
	newParagraph = "\n\n"
)

func ToString(comments endpoints.Comments, config *core.Config, screenWidth int) string {
	header := getHeader(comments, config, screenWidth)

	replies := ""

	for _, reply := range comments.Comments {
		replies += printReplies(reply, config, screenWidth, comments.User, "")
	}

	return header + replies
}

func getHeader(c endpoints.Comments, config *core.Config, screenWidth int) string {
	if config.CommentWidth == 0 {
		config.CommentWidth = screenWidth
	}

	headline := getHeadline(c.Title, config)
	infoLine := getInfoLine(c.Points, c.User, c.TimeAgo, c.CommentsCount, c.ID)
	helpMessage := aurora.Faint(messages.LessScreenInfo).Faint().String() + newLine
	helpMessage += aurora.Faint(messages.LessCommentInfo).Faint().String() + newLine
	url := getURL(c.URL, c.Domain, config)
	rootComment := parseRootComment(c.Content, config)
	separator := messages.GetSeparator(config.CommentWidth)

	return headline + helpMessage + newLine + infoLine + url + rootComment + separator + newParagraph
}

func getHeadline(title string, config *core.Config) string {
	formattedTitle := highlightTitle(unicode.ZeroWidthSpace+title, config.HighlightHeadlines)
	wrappedHeadline, _ := text.Wrap(formattedTitle, config.CommentWidth)

	return wrappedHeadline + newParagraph
}

func getURL(url string, domain string, config *core.Config) string {
	if domain == "" {
		url = "https://news.ycombinator.com/" + url
	}

	truncatedURL := text.TruncateMax(url, config.CommentWidth)
	formattedURL := aurora.Faint(truncatedURL).String() + newLine

	return formattedURL
}

func highlightTitle(title string, highlightHeadlines bool) string {
	if highlightHeadlines {
		title = syntax.HighlightYCStartups(title)
		title = syntax.HighlightHackerNewsHeadlines(title)
		title = syntax.HighlightSpecialContent(title)
	}

	return title
}

func getInfoLine(points int, user string, timeAgo string, numberOfComments int, id int) string {
	p := strconv.Itoa(points)
	c := strconv.Itoa(numberOfComments)
	i := strconv.Itoa(id)

	formattedInfoLine := aurora.Faint(p + " points by " + user + " " + timeAgo +
		" • " + c + " comments" + " • " + "ID " + i).String()

	return formattedInfoLine + newLine
}

func parseRootComment(c string, config *core.Config) string {
	if c == "" {
		return ""
	}

	comment := ParseComment(c, config, config.CommentWidth, config.CommentWidth)
	wrappedComment, _ := text.Wrap(comment, config.CommentWidth)

	return newLine + wrappedComment + newLine
}

func printReplies(c endpoints.Comments, config *core.Config, screenWidth int, originalPoster string,
	parentPoster string) string {
	isDeletedAndHasNoReplies := c.Content == "[deleted]" && len(c.Comments) == 0
	if isDeletedAndHasNoReplies {
		return ""
	}

	currentIndentSize := config.IndentSize * c.Level
	usableScreenSize := screenWidth - currentIndentSize - 1
	adjustedCommentWidth := getCommentWidthForLevel(currentIndentSize, usableScreenSize, config.CommentWidth,
		config.PreserveRightMargin)

	comment := ParseComment(c.Content, config, adjustedCommentWidth, usableScreenSize)

	indentSymbol := indent.GetIndentSymbol(config.HideIndentSymbol, config.AltIndentBlock)
	indentBlock := getIndentBlock(indentSymbol, c.Level, config.IndentSize)
	paddingWithBlock := text.WrapPad(indentBlock)
	wrappedAndPaddedComment, _ := text.Wrap(comment, screenWidth, paddingWithBlock)

	paddingWithNoBlock := text.WrapPad(getIndentBlockWithoutBar(c.Level, config.IndentSize))

	author := getCommentHeading(c, c.Level, config.CommentWidth, originalPoster, parentPoster)
	paddedAuthor, _ := text.Wrap(author, usableScreenSize, paddingWithNoBlock)
	fullComment := paddedAuthor + wrappedAndPaddedComment + newParagraph

	if c.Level == 0 {
		parentPoster = c.User
	}

	for _, reply := range c.Comments {
		fullComment += printReplies(reply, config, screenWidth, originalPoster, parentPoster)
	}

	return fullComment
}

func getCommentHeading(c endpoints.Comments, level int, commentWidth int, originalPoster string,
	parentPoster string) string {
	timeAgo := c.TimeAgo
	author := aurora.Bold(c.User).String()
	label := getAuthorLabel(c.User, originalPoster, parentPoster) + " "

	if level == 0 {
		replies := getRepliesTag(getReplyCount(c))
		anchor := unicode.ZeroWidthSpace
		lengthOfUnderline := commentWidth - text.Len(author+label+anchor+timeAgo+replies)
		headerLine := strings.Repeat(" ", lengthOfUnderline)
		info := aurora.Faint(timeAgo + headerLine + replies + anchor).Underline().String()

		return author + label + info + newLine
	}

	return author + label + aurora.Faint(timeAgo).String() + newLine
}

func getRepliesTag(numberOfReplies int) string {
	if numberOfReplies == 0 {
		return ""
	}

	return strconv.Itoa(numberOfReplies) + " ⤶"
}

func getReplyCount(comments endpoints.Comments) int {
	numberOfReplies := 0

	return incrementReplyCount(comments, &numberOfReplies)
}

func incrementReplyCount(comments endpoints.Comments, repliesSoFar *int) int {
	for _, reply := range comments.Comments {
		*repliesSoFar++
		incrementReplyCount(reply, repliesSoFar)
	}

	return *repliesSoFar
}

// Adjusted comment width shortens the commentWidth if the available screen size
// is smaller than the size of the commentWidth.
func getCommentWidthForLevel(currentIndentSize int, usableScreenSize int, commentWidth int,
	preserveRightMargin bool) int {
	if usableScreenSize < commentWidth || commentWidth == 0 {
		return usableScreenSize
	}

	if preserveRightMargin {
		return commentWidth - currentIndentSize
	}

	return commentWidth
}

func getAuthorLabel(author, originalPoster, parentPoster string) string {
	switch author {
	case "dang":
		return aurora.Green(" mod").String()

	case originalPoster:
		return aurora.Red(" OP").String()

	case parentPoster:
		return aurora.Magenta(" PP").String()

	default:
		return ""
	}
}

func getIndentBlockWithoutBar(level int, indentSize int) string {
	if level == 0 {
		return ""
	}

	return strings.Repeat(" ", indentSize*level+1)
}

func getIndentBlock(indentSymbol string, level int, indentSize int) string {
	if level == 0 {
		return ""
	}

	// indentation := colors.Normal + colors.GetIndentBlockColor(level) + indentSymbol + colors.Normal
	indentation := syntax.ColorizeIndentSymbol(indentSymbol, level)
	whitespace := strings.Repeat(" ", indentSize*level)

	return whitespace + indentation
}
