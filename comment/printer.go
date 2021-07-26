package comment

import (
	"clx/colors"
	"clx/constants/messages"
	"clx/core"
	"clx/endpoints"
	"clx/syntax"
	"strconv"
	"strings"

	text "github.com/MichaelMure/go-term-text"
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

	headline := getHeadline(c.Title, c.User, c.Domain, c.URL, c.ID, config.CommentWidth)
	infoLine := getInfoLine(c.Points, c.User, c.TimeAgo, c.CommentsCount, c.ID)
	rootComment := parseRootComment(c.Content, config)
	helpMessage := colors.ToDimmed(messages.LessScreenInfo) + colors.NewLine
	separator := messages.GetSeparator(config.CommentWidth)

	return headline + infoLine + helpMessage + rootComment + separator + colors.NewParagraph
}

func getHeadline(title, author, domain, url string, id, commentWidth int) string {
	if domain == "" {
		domain = "item?id=" + strconv.Itoa(id)
	}

	// TODO: disable this when syntax highlighting for headlines is disabled
	title = highlightTitle(title, author)
	headline := title + " " + colors.SurroundWithParen(domain)
	wrappedHeadline, _ := text.Wrap(headline, commentWidth)
	hyperlink := getHyperlink(domain, url, id)

	wrappedHeadline = strings.ReplaceAll(wrappedHeadline, domain, hyperlink)

	return wrappedHeadline + colors.NewLine
}

func highlightTitle(title, author string) string {
	if author == "whoishiring" {
		return syntax.HighlightWhoIsHiring(title, author)
	}

	title = syntax.HighlightYCStartups(title)
	title = syntax.HighlightHackerNewsHeadlines(title)

	return title
}

func getHyperlink(domain string, url string, id int) string {
	if domain != "" {
		return getHyperlinkText(url, domain)
	}

	linkToComments := "https://news.ycombinator.com/item?id=" + strconv.Itoa(id)
	linkText := "item?id=" + strconv.Itoa(id)

	return getHyperlinkText(linkToComments, linkText)
}

func getInfoLine(points int, user string, timeAgo string, numberOfComments int, id int) string {
	p := strconv.Itoa(points)
	c := strconv.Itoa(numberOfComments)
	i := strconv.Itoa(id)

	return colors.ToDimmed(p+" points by "+user+" "+timeAgo+" • "+c+" comments"+" • "+"ID "+i) + colors.NewLine
}

func getHyperlinkText(url string, text string) string {
	return colors.Link1 + url + colors.Link2 + text + colors.Link3
}

func parseRootComment(c string, config *core.Config) string {
	if c == "" {
		return ""
	}

	comment := ParseComment(c, config, config.CommentWidth, config.CommentWidth)
	wrappedComment, _ := text.Wrap(comment, config.CommentWidth)

	return colors.NewLine + wrappedComment + colors.NewLine
}

func printReplies(c endpoints.Comments, config *core.Config, screenWidth int, originalPoster string,
	parentPoster string) string {
	isDeletedAndHasNoReplies := c.Content == "[deleted]" && len(c.Comments) == 0
	if isDeletedAndHasNoReplies {
		return ""
	}

	currentIndentSize := config.IndentSize * c.Level
	usableScreenSize := screenWidth - currentIndentSize - 1
	adjustedCommentWidth := getCommentWidthForLevel(currentIndentSize, usableScreenSize, c.Level, config.IndentSize,
		config.CommentWidth, config.PreserveRightMargin)

	comment := ParseComment(c.Content, config, adjustedCommentWidth, usableScreenSize)

	indentBlock := getIndentBlock(c.Level, config.IndentSize, config.AltIndentBlock)
	paddingWithBlock := text.WrapPad(indentBlock)
	wrappedAndPaddedComment, _ := text.Wrap(comment, screenWidth, paddingWithBlock)

	paddingWithNoBlock := text.WrapPad(getIndentBlockWithoutBar(c.Level, config.IndentSize))

	author := getCommentHeading(c, c.Level, config.CommentWidth, originalPoster, parentPoster)
	paddedAuthor, _ := text.Wrap(author, adjustedCommentWidth, paddingWithNoBlock)
	fullComment := paddedAuthor + wrappedAndPaddedComment + colors.NewParagraph

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
	author := colors.ToBold(c.User)
	label := getAuthorLabel(c.User, originalPoster, parentPoster) + " "

	if level == 0 {
		replies := getRepliesTag(getReplyCount(c))
		anchor := " ::"
		lengthOfUnderline := commentWidth - text.Len(author+label+anchor+timeAgo+replies)
		headerLine := strings.Repeat(" ", lengthOfUnderline)

		return author + label + colors.ToDimmedAndUnderlined(timeAgo+headerLine+replies+anchor) + colors.NewLine
	}

	return author + label + colors.ToDimmed(timeAgo) + colors.NewLine
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
func getCommentWidthForLevel(currentIndentSize int, usableScreenSize int, level int, indentSize int, commentWidth int,
	preserveRightMargin bool) int {
	if usableScreenSize < commentWidth || commentWidth == 0 {
		return usableScreenSize
	}

	if preserveRightMargin {
		return commentWidth - currentIndentSize
	}

	// return commentWidth + indentSize*level
	return commentWidth
}

func getAuthorLabel(author, originalPoster, parentPoster string) string {
	switch author {
	// case "":
	//	return ""
	case "dang":
		return colors.ToGreen(" mod")
	case originalPoster:
		return colors.ToRed(" OP")
	case parentPoster:
		return colors.ToMagenta(" PP")
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

func getIndentBlock(level int, indentSize int, altIndentBlock bool) string {
	if level == 0 {
		return ""
	}

	indentBlock := getIndentationSymbol(altIndentBlock)
	indentation := colors.Normal + colors.GetIndentBlockColor(level) + indentBlock + colors.Normal
	whitespace := strings.Repeat(" ", indentSize*level)

	return whitespace + indentation
}

func getIndentationSymbol(useAlternateIndent bool) string {
	if useAlternateIndent {
		return "┃"
	}

	return "▎"
}
