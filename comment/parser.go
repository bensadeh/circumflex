package comment

import (
	"clx/screen"
	"regexp"
	"strconv"
	"strings"

	text "github.com/MichaelMure/go-term-text"
)

func PrintCommentTree(comments Comments, indentSize int, commentWidth int, screenWidth int,
	preserveRightMargin bool) string {
	header := getHeader(comments, commentWidth)
	originalPoster := comments.User
	commentTree := ""

	for _, reply := range comments.Comments {
		commentTree += prettyPrintComments(reply, indentSize, commentWidth, screenWidth,
			originalPoster, "", preserveRightMargin)
	}

	return header + commentTree
}

func getHeader(c Comments, commentWidth int) string {
	if commentWidth == 0 {
		commentWidth = screen.GetTerminalWidth()
	}

	headline := getHeadline(c.Title, c.Domain, c.URL, c.ID, commentWidth)
	infoLine := getInfoLine(c.Points, c.User, c.TimeAgo, c.CommentsCount)
	submissionComment := parseRootComment(c.Content, commentWidth)
	helpMessage := dimmed("You are now in 'less'. Press 'q' to return and 'h' for help.") + NewLine
	separator := strings.Repeat("-", commentWidth)

	return headline + infoLine + helpMessage + submissionComment + separator + NewParagraph
}

func getInfoLine(points int, author string, timeAgo string, numberOfComments int) string {
	p := strconv.Itoa(points)
	c := strconv.Itoa(numberOfComments)

	return dimmed(p+" points by "+author+" "+timeAgo+" • "+c+" comments") + NewLine
}

func getHeadline(title, domain, url string, id, commentWidth int) string {
	if domain == "" {
		domain = "item?id=" + strconv.Itoa(id)
	}

	headline := title + " " + paren(domain) + NewLine
	wrappedHeadline, _ := text.Wrap(headline, commentWidth)
	hyperlink := getHyperlink(domain, url, id)

	wrappedHeadline = strings.ReplaceAll(wrappedHeadline, domain, hyperlink)

	return wrappedHeadline
}

func getHyperlink(domain string, URL string, id int) string {
	if domain != "" {
		return getHyperlinkText(URL, domain)
	}

	linkToComments := "https://news.ycombinator.com/item?id=" + strconv.Itoa(id)
	linkText := "item?id=" + strconv.Itoa(id)

	return getHyperlinkText(linkToComments, linkText)
}

func getHyperlinkText(url string, text string) string {
	return Link1 + url + Link2 + text + Link3
}

func parseRootComment(c string, commentWidth int) string {
	if c == "" {
		return ""
	}

	comment, URLs := parseComment(c)
	wrappedComment, _ := text.Wrap(comment, commentWidth)
	wrappedComment = applyURLs(wrappedComment, URLs)

	return NewLine + wrappedComment + NewLine
}

func prettyPrintComments(c Comments, indentSize int, commentWidth int, screenWidth int, originalPoster string,
	parentPoster string, preserveRightMargin bool) string {
	comment, URLs := parseComment(c.Content)
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
		fullComment += prettyPrintComments(reply, indentSize, commentWidth, screenWidth,
			originalPoster, parentPoster, preserveRightMargin)
	}

	return fullComment
}

func getCommentHeading(c Comments, level int, commentWidth int, originalPoster string, parentPoster string) string {
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

func getReplyCount(comments Comments) int {
	numberOfReplies := 0

	return incrementReplyCount(comments, &numberOfReplies)
}

func incrementReplyCount(comments Comments, repliesSoFar *int) int {
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
// is smaller than the size of the commentWidth
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

	indentation := " "

	for i := 0; i < indentSize*level; i++ {
		indentation += " "
	}

	return indentation
}

func getIndentBlock(level int, indentSize int) string {
	if level == 0 {
		return ""
	}

	indentation := Normal + getColoredIndentBlock(level) + "▎" + Normal
	whitespace := strings.Repeat(" ", indentSize*level)

	return whitespace + indentation
}

func parseComment(comment string) (string, []string) {
	comment = replaceCharacters(comment)
	comment = replaceHTML(comment)
	comment = colorizeLinkNumbers(comment)
	URLs := extractURLs(comment)
	comment = trimURLs(comment)

	return comment, URLs
}

func replaceCharacters(input string) string {
	input = strings.ReplaceAll(input, "&#x27;", "'")
	input = strings.ReplaceAll(input, "&gt;", ">")
	input = strings.ReplaceAll(input, "&lt;", "<")
	input = strings.ReplaceAll(input, "&#x2F;", "/")
	input = strings.ReplaceAll(input, "&quot;", "\"")
	input = strings.ReplaceAll(input, "&amp;", "&")
	input = strings.ReplaceAll(input, ".  ", ". ")
	input = strings.ReplaceAll(input, "!  ", "! ")
	input = strings.ReplaceAll(input, "?  ", "? ")

	return input
}

func replaceHTML(input string) string {
	input = strings.Replace(input, "<p>", "", 1)

	input = strings.ReplaceAll(input, "<p>", NewParagraph)
	input = strings.ReplaceAll(input, "<i>", Italic)
	input = strings.ReplaceAll(input, "</i>", Normal)
	input = strings.ReplaceAll(input, "</a>", "")
	input = strings.ReplaceAll(input, "<pre><code>", Dimmed)
	input = strings.ReplaceAll(input, "</code></pre>", Normal)

	return input
}

func colorizeLinkNumbers(input string) string {
	input = strings.ReplaceAll(input, "[0]", "["+white("0")+"]")
	input = strings.ReplaceAll(input, "[1]", "["+red("1")+"]")
	input = strings.ReplaceAll(input, "[2]", "["+yellow("2")+"]")
	input = strings.ReplaceAll(input, "[3]", "["+green("3")+"]")
	input = strings.ReplaceAll(input, "[4]", "["+blue("4")+"]")
	input = strings.ReplaceAll(input, "[5]", "["+cyan("5")+"]")
	input = strings.ReplaceAll(input, "[6]", "["+magenta("6")+"]")
	input = strings.ReplaceAll(input, "[7]", "["+altWhite("7")+"]")
	input = strings.ReplaceAll(input, "[8]", "["+altRed("8")+"]")
	input = strings.ReplaceAll(input, "[9]", "["+altYellow("9")+"]")
	input = strings.ReplaceAll(input, "[10]", "["+altGreen("10")+"]")

	return input
}

func extractURLs(input string) []string {
	expForFirstTag := regexp.MustCompile(`<a href=".*?" rel="nofollow">`)
	URLs := expForFirstTag.FindAllString(input, 10)

	for i := range URLs {
		URLs[i] = strings.ReplaceAll(URLs[i], `<a href="`, "")
		URLs[i] = strings.ReplaceAll(URLs[i], `" rel="nofollow">`, "")
	}

	return URLs
}

func trimURLs(comment string) string {
	expression := regexp.MustCompile(`<a href=".*?" rel="nofollow">`)

	return expression.ReplaceAllString(comment, "")
}
