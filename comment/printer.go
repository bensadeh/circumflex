package comment

import (
	"clx/constants/messages"
	"clx/endpoints"
	"strconv"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

func ToString(comments endpoints.Comments, indentSize int, commentWidth int, screenWidth int, preserveRightMargin bool,
	altIndentBlock bool, commentHighlighting bool, emojiSmiley bool) string {
	header := getHeader(comments, commentWidth, screenWidth, commentHighlighting, altIndentBlock, emojiSmiley)
	replies := ""

	for _, reply := range comments.Comments {
		replies += printReplies(reply, indentSize, commentWidth, screenWidth, comments.User, "",
			preserveRightMargin, altIndentBlock, commentHighlighting, emojiSmiley)
	}

	return header + replies
}

func getHeader(c endpoints.Comments, commentWidth int, screenWidth int, commentHighlighting bool,
	altIndentBlock bool, emojiSmiley bool) string {
	if commentWidth == 0 {
		commentWidth = screenWidth
	}

	headline := getHeadline(c.Title, c.Domain, c.URL, c.ID, commentWidth)
	infoLine := getInfoLine(c.Points, c.User, c.TimeAgo, c.CommentsCount, c.ID)
	rootComment := parseRootComment(c.Content, commentWidth, commentHighlighting, altIndentBlock, emojiSmiley)
	helpMessage := dimmed(messages.LessScreenInfo) + NewLine
	separator := messages.GetSeparator(commentWidth)

	return headline + infoLine + helpMessage + rootComment + separator + NewParagraph
}

func getHeadline(title, domain, url string, id, commentWidth int) string {
	if domain == "" {
		domain = "item?id=" + strconv.Itoa(id)
	}

	headline := title + " " + paren(domain)
	wrappedHeadline, _ := text.Wrap(headline, commentWidth)
	hyperlink := getHyperlink(domain, url, id)

	wrappedHeadline = strings.ReplaceAll(wrappedHeadline, domain, hyperlink)

	return wrappedHeadline + NewLine
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

	return dimmed(p+" points by "+user+" "+timeAgo+" • "+c+" comments"+" • "+"ID "+i) + NewLine
}

func getHyperlinkText(url string, text string) string {
	return Link1 + url + Link2 + text + Link3
}

func parseRootComment(c string, commentWidth int, commentHighlighting bool, altIndentBlock bool,
	emojiSmiley bool) string {
	if c == "" {
		return ""
	}

	comment := ParseComment(c, commentWidth, commentWidth, commentHighlighting, altIndentBlock, emojiSmiley)
	wrappedComment, _ := text.Wrap(comment, commentWidth)

	return NewLine + wrappedComment + NewLine
}

func printReplies(c endpoints.Comments, indentSize int, commentWidth int, screenWidth int, originalPoster string,
	parentPoster string, preserveRightMargin bool, altIndentBlock bool, commentHighlighting bool,
	emojiSmiley bool) string {
	currentIndentSize := indentSize * c.Level
	usableScreenSize := screenWidth - currentIndentSize
	comment := ParseComment(c.Content, commentWidth, usableScreenSize, commentHighlighting, altIndentBlock, emojiSmiley)
	adjustedCommentWidth := getCommentWidthForLevel(c.Level, indentSize, commentWidth, screenWidth, preserveRightMargin)

	indentBlock := getIndentBlock(c.Level, indentSize, altIndentBlock)
	paddingWithBlock := text.WrapPad(indentBlock)
	wrappedAndPaddedComment, _ := text.Wrap(comment, screenWidth, paddingWithBlock)

	paddingWithNoBlock := text.WrapPad(getIndentBlockWithoutBar(c.Level, indentSize))

	author := getCommentHeading(c, c.Level, commentWidth, originalPoster, parentPoster)
	paddedAuthor, _ := text.Wrap(author, adjustedCommentWidth, paddingWithNoBlock)
	fullComment := paddedAuthor + wrappedAndPaddedComment + NewParagraph

	if c.Level == 0 {
		parentPoster = c.User
	}

	for _, reply := range c.Comments {
		fullComment += printReplies(reply, indentSize, commentWidth, screenWidth,
			originalPoster, parentPoster, preserveRightMargin, altIndentBlock, commentHighlighting, emojiSmiley)
	}

	return fullComment
}

func getCommentHeading(c endpoints.Comments, level int, commentWidth int, originalPoster string,
	parentPoster string) string {
	timeAgo := c.TimeAgo
	author := bold(c.User)
	label := getAuthorLabel(c.User, originalPoster, parentPoster) + " "

	if level == 0 {
		replies := getRepliesTag(getReplyCount(c))
		anchor := " ::"
		lengthOfUnderline := commentWidth - text.Len(author+label+anchor+timeAgo+replies)
		headerLine := strings.Repeat(" ", lengthOfUnderline)

		return author + label + dimmedAndUnderlined(timeAgo+headerLine+replies+anchor) + NewLine
	}

	return author + label + dimmed(timeAgo) + NewLine
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
func getCommentWidthForLevel(level int, indentSize int, commentWidth int, screenWidth int,
	preserveRightMargin bool) int {
	currentIndentSize := indentSize * level
	usableScreenSize := screenWidth - currentIndentSize

	if usableScreenSize < commentWidth || commentWidth == 0 {
		return usableScreenSize + currentIndentSize
	}

	if preserveRightMargin {
		return commentWidth
	}

	return commentWidth + indentSize*level
}

func getAuthorLabel(author, originalPoster, parentPoster string) string {
	switch author {
	case "":
		return ""
	case "dang":
		return green(" mod")
	case originalPoster:
		return red(" OP")
	case parentPoster:
		return magenta(" PP")
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
	indentation := Normal + getColoredIndentBlock(level) + indentBlock + Normal
	whitespace := strings.Repeat(" ", indentSize*level)

	return whitespace + indentation
}

func getIndentationSymbol(useAlternateIndent bool) string {
	if useAlternateIndent {
		return "┃"
	}

	return "▎"
}
