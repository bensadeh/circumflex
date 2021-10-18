package comment

import (
	"clx/comment/postprocessor"
	"clx/constants/margins"
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
	commentSectionScreenWidth := screenWidth - margins.CommentSectionLeftMargin

	header := getHeader(comments, config)

	replies := ""

	for _, reply := range comments.Comments {
		replies += printReplies(reply, config, commentSectionScreenWidth, comments.User, "")
	}

	commentSection := postprocessor.Process(header+replies, screenWidth)

	return commentSection
}

func getHeader(c endpoints.Comments, config *core.Config) string {
	headline := getHeadline(c.Title, config)
	infoLine := getInfoLine(c.Points, c.User, c.TimeAgo, c.CommentsCount, c.ID)
	helpMessage := aurora.Faint(messages.LessScreenInfo).Faint().String() + newLine
	helpMessage += aurora.Faint(messages.LessCommentInfo).Faint().String() + newLine
	url := getURL(c.URL, c.Domain, config)
	rootComment := parseRootComment(c.Content, config)
	separator := aurora.Yellow(messages.GetSeparator(config.CommentWidth)).String()

	return headline + separator + newLine + helpMessage + newLine + infoLine + url + rootComment + separator + newParagraph
}

func getHeadline(title string, config *core.Config) string {
	formattedTitle := highlightTitle(unicode.ZeroWidthSpace+newLine+title, config.HighlightHeadlines)
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
	highlightedTitle := ""

	if highlightHeadlines {
		highlightedTitle = syntax.HighlightYCStartupsInHeadlines(title)
		highlightedTitle = syntax.HighlightHackerNewsHeadlines(highlightedTitle)
		highlightedTitle = syntax.HighlightSpecialContent(highlightedTitle)
	}

	return aurora.Bold(highlightedTitle).String()
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

	indentation := getIndentString(c.Level)
	indentSize := len(indentation)
	availableScreenWidth := screenWidth - indentSize - margins.CommentSectionLeftMargin
	adjustedCommentWidth := config.CommentWidth - c.Level

	comment := formatComment(c, config, originalPoster, parentPoster, adjustedCommentWidth, availableScreenWidth)
	indentedComment, _ := text.WrapWithPad(comment, availableScreenWidth, indentation)
	fullComment := indentedComment + newParagraph

	if c.Level == 0 {
		parentPoster = c.User
	}

	for _, reply := range c.Comments {
		fullComment += printReplies(reply, config, screenWidth, originalPoster, parentPoster)
	}

	return fullComment
}

func formatComment(c endpoints.Comments, config *core.Config, originalPoster string, parentPoster string,
	commentWidth int, availableScreenWidth int) string {
	indentSymbol := indent.GetIndentSymbol(config.HideIndentSymbol, config.AltIndentBlock)
	coloredIndentSymbol := syntax.ColorizeIndentSymbol(indentSymbol, c.Level)

	header := getCommentHeader(c, originalPoster, parentPoster, commentWidth)
	comment := ParseComment(c.Content, config, commentWidth, availableScreenWidth)

	paddedComment, _ := text.WrapWithPad(comment, availableScreenWidth, coloredIndentSymbol)

	return header + paddedComment
}

func getIndentString(level int) string {
	if level == 0 {
		return ""
	}

	return strings.Repeat(" ", level-1)
}

func getCommentHeader(c endpoints.Comments, originalPoster string, parentPoster string, commentWidth int) string {
	if c.Level == 0 {
		return formatHeader(c, originalPoster, parentPoster, true, true, true,
			commentWidth, 0)
	}

	return formatHeader(c, originalPoster, parentPoster, false, false, false,
		commentWidth, 1)
}

func formatHeader(c endpoints.Comments, originalPoster string, parentPoster string, underlineHeader bool,
	showReplies bool, enableZeroWidthSpace bool, commentWidth int, indentSize int) string {
	authorInBold := aurora.Bold(c.User).String() + " "
	authorLabel := getAuthorLabel(c.User, originalPoster, parentPoster)
	zeroWidthSpace := getZeroWidthSpace(enableZeroWidthSpace)
	repliesTag := getReplies(showReplies, c)
	indentation := strings.Repeat(" ", indentSize)

	spacingLength := commentWidth - text.Len(indentation+authorInBold+authorLabel+c.TimeAgo+repliesTag)
	spacing := strings.Repeat(" ", spacingLength)

	return zeroWidthSpace + indentation + authorInBold + authorLabel +
		underlineAndDim(underlineHeader, c.TimeAgo, spacing, repliesTag) + newLine
}

func underlineAndDim(enabled bool, timeAgo, spacing, replies string) string {
	if enabled {
		return aurora.Underline(timeAgo + spacing + replies).Faint().String()
	}

	return aurora.Faint(timeAgo).String()
}

func getReplies(showReplies bool, children endpoints.Comments) string {
	if !showReplies {
		return ""
	}

	numberOfReplies := getReplyCount(children)
	replyTag := getRepliesTag(numberOfReplies)

	return replyTag
}

func getRepliesTag(numberOfReplies int) string {
	if numberOfReplies == 0 {
		return ""
	}

	return strconv.Itoa(numberOfReplies) + " ↩"
}

func getZeroWidthSpace(enabled bool) string {
	if enabled {
		return unicode.ZeroWidthSpace
	}

	return ""
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

func getAuthorLabel(author, originalPoster, parentPoster string) string {
	switch author {
	case "dang":
		return aurora.Green("mod ").String()

	case originalPoster:
		return aurora.Red("OP ").String()

	case parentPoster:
		return aurora.Magenta("PP ").String()

	default:
		return ""
	}
}
