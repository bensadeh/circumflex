package comment

import (
	"clx/endpoints"
	"strconv"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

func ToString(comments endpoints.Comments, indentSize int, commentWidth int, screenWidth int, preserveRightMargin bool) string {
	header := getHeader(comments, commentWidth, screenWidth)
	replies := ""

	for _, reply := range comments.Comments {
		replies += printReplies(reply, indentSize, commentWidth, screenWidth, comments.User, "",
			preserveRightMargin)
	}

	return header + replies
}

func getHeader(c endpoints.Comments, commentWidth int, screenWidth int) string {
	if commentWidth == 0 {
		commentWidth = screenWidth
	}

	headline := getHeadline(c.Title, c.Domain, c.URL, c.ID, commentWidth)
	infoLine := getInfoLine(c.Points, c.User, c.TimeAgo, c.CommentsCount)
	rootComment := parseRootComment(c.Content, commentWidth)
	helpMessage := dimmed("You are now in 'less'. Press 'q' to return and 'h' for help.") + NewLine
	separator := strings.Repeat("-", commentWidth)

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

func getInfoLine(points int, user string, timeAgo string, numberOfComments int) string {
	p := strconv.Itoa(points)
	c := strconv.Itoa(numberOfComments)

	return dimmed(p+" points by "+user+" "+timeAgo+" • "+c+" comments") + NewLine
}

func getHyperlinkText(url string, text string) string {
	return Link1 + url + Link2 + text + Link3
}

func parseRootComment(c string, commentWidth int) string {
	if c == "" {
		return ""
	}

	comment, URLs := ParseComment(c)
	wrappedComment, _ := text.Wrap(comment, commentWidth)
	wrappedComment = applyURLs(wrappedComment, URLs)

	return NewLine + wrappedComment + NewLine
}

func printReplies(c endpoints.Comments, indentSize int, commentWidth int, screenWidth int, originalPoster string,
	parentPoster string, preserveRightMargin bool) string {
	comment, URLs := ParseComment(c.Content)
	adjustedCommentWidth := getCommentWidthForLevel(c.Level, indentSize, commentWidth, screenWidth, preserveRightMargin)

	indentBlock := getIndentBlock(c.Level, indentSize)
	paddingWithBlock := text.WrapPad(indentBlock)
	wrappedAndPaddedComment, _ := text.Wrap(comment, adjustedCommentWidth, paddingWithBlock)

	paddingWithNoBlock := text.WrapPad(getIndentBlockWithoutBar(c.Level, indentSize))

	author := getCommentHeading(c, c.Level, commentWidth, originalPoster, parentPoster)
	paddedAuthor, _ := text.Wrap(author, adjustedCommentWidth, paddingWithNoBlock)
	fullComment := paddedAuthor + wrappedAndPaddedComment + NewParagraph
	fullComment = applyURLs(fullComment, URLs)

	if c.Level == 0 {
		parentPoster = c.User
	}

	for _, reply := range c.Comments {
		fullComment += printReplies(reply, indentSize, commentWidth, screenWidth,
			originalPoster, parentPoster, preserveRightMargin)
	}

	return fullComment
}

func getCommentHeading(c endpoints.Comments, level int, commentWidth int, originalPoster string, parentPoster string) string {
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

func applyURLs(comment string, urls []string) string {
	for _, url := range urls {
		truncatedURL := truncateURL(url)
		URLWithHyperlinkCode := getHyperlinkText(url, truncatedURL)
		comment = strings.ReplaceAll(comment, truncatedURL, URLWithHyperlinkCode)
	}

	return comment
}

func truncateURL(url string) string {
	const hackerNewsMaxURLLength = 60

	if len(url) < hackerNewsMaxURLLength {
		return url
	}

	truncatedURL := ""

	for i, c := range url {
		if i == hackerNewsMaxURLLength {
			truncatedURL += "..."

			break
		}

		truncatedURL += string(c)
	}

	return truncatedURL
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

func getIndentBlock(level int, indentSize int) string {
	if level == 0 {
		return ""
	}

	indentation := Normal + getColoredIndentBlock(level) + "▎" + Normal
	whitespace := strings.Repeat(" ", indentSize*level)

	return whitespace + indentation
}
