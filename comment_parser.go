package main

import (
	text "github.com/MichaelMure/go-term-text"
	"regexp"
	"strconv"
	"strings"

	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

// Comments represent the JSON structure as
// retrieved from cheeaun's unofficial HN API
type Comments struct {
	Author        string      `json:"user"`
	Title         string      `json:"title"`
	Comment       string      `json:"content"`
	CommentsCount int         `json:"comments_count"`
	Time          string      `json:"time_ago"`
	Points        int         `json:"points"`
	URL           string      `json:"url"`
	Domain        string      `json:"domain"`
	ID            int         `json:"id"`
	Replies       []*Comments `json:"comments"`
}

func printCommentTree(comments Comments, indentSize int, commentWith int) string {
	header := getHeader(comments, commentWith)
	originalPoster := comments.Author
	commentTree := ""
	for _, reply := range comments.Replies {
		commentTree += prettyPrintComments(*reply, 0, indentSize, commentWith, originalPoster, "")
	}
	return header + commentTree
}

func getHeader(c Comments, commentWidth int) string {
	headline := c.Title + getDomainText(c.Domain, c.URL, c.ID) + NewLine
	infoLine := dimmed(strconv.Itoa(c.Points)+" points by "+c.Author+" "+c.Time+" • "+strconv.Itoa(c.CommentsCount)+" comments") + NewLine

	header := headline + infoLine
	header += parseRootComment(c.Comment, commentWidth)

	for i := 0; i < commentWidth; i++ {
		header += "-"
	}

	return header + DoubleNewLine
}

func getDomainText(domain string, URL string, id int) string {
	if domain != "" {
		return " " + paren(getHyperlinkText(URL, domain))
	}
	linkToComments := "https://news.ycombinator.com/item?id=" + strconv.Itoa(id)
	linkText := "item?id=" + strconv.Itoa(id)
	return " " + paren(getHyperlinkText(linkToComments, linkText))
}

func getHyperlinkText(URL string, text string) string {
	return Link1 + URL + Link2 + text + Link3
}

func parseRootComment(comment string, lineLength int) string {
	if comment == "" {
		return ""
	}

	//parsedComment := parseComment(comment)
	//
	//commentLines := strings.Split(parsedComment, "<p>")
	//lastParagraph := len(commentLines) - 1
	//firstParagraph := 0
	fullComment := ""
	//for i, line := range commentLines {
	//	wrapped := wordwrap.WrapString(line, uint(lineLength))
	//	wrappedAndIndentedComment := wordwrap.Indent(wrapped, getIndentBlock(0, 0), true)
	//	if i == firstParagraph {
	//		fullComment = NewLine
	//	}
	//	if i == lastParagraph {
	//		fullComment += wrappedAndIndentedComment + NewLine
	//	} else {
	//		fullComment += wrappedAndIndentedComment + DoubleNewLine
	//	}
	//}
	return fullComment
}

func prettyPrintComments(c Comments, level int, indentSize int, commentWidth int, originalPoster string, parentPoster string) string {
	comment, URLs := parseComment(c.Comment)
	adjustedCommentWidth := getAdjustedCommentWidth(level, indentSize, commentWidth)

	indentBlock := getIndentBlock(level, indentSize)
	paddingWithBlock := text.WrapPad(indentBlock)
	wrappedAndPaddedComment, _ := text.Wrap(comment, adjustedCommentWidth+indentSize, paddingWithBlock)

	paddingWithNoBlock := text.WrapPad(getIndentBlockWithoutBar(level, indentSize))
	author := labelAuthor(c.Author, originalPoster, parentPoster) + " " + dimmed(c.Time) + getTopLevelCommentAnchor(level) + NewLine
	paddedAuthor, _ := text.Wrap(author, commentWidth, paddingWithNoBlock)
	fullComment := paddedAuthor + wrappedAndPaddedComment + DoubleNewLine
	fullComment = applyURLs(fullComment, URLs)

	if level == 0 {
		parentPoster = c.Author
	}

	for _, s := range c.Replies {
		fullComment += prettyPrintComments(*s, level+1, indentSize, commentWidth, originalPoster, parentPoster)
	}
	return fullComment
}

func applyURLs(comment string, URLs []string) string {
	for _, URL := range URLs {
		truncatedURL := truncateURL(URL)
		URLWithHyperlinkCode := getHyperlinkText(URL, truncatedURL)
		comment = strings.ReplaceAll(comment, truncatedURL, URLWithHyperlinkCode)
	}
	return comment
}

func truncateURL(URL string) string {
	if len(URL) < 60 {
		return URL
	}

	truncatedURL := ""
	for i, c := range URL {
		if i == 60 {
			truncatedURL += "..."
			break
		}
		truncatedURL += string(c)
	}
	return truncatedURL
}

func getTopLevelCommentAnchor(level int) string {
	if level == 0 {
		return dimmed(" ::")
	}
	return ""
}

// Adjusted comment width shortens the commentWidth if the available screen size
// is smaller than the size of the commentWidth
func getAdjustedCommentWidth(level int, indentSize int, commentWidth int) int {
	x, _ := terminal.Width()
	screenWidth := int(x)

	currentIndentSize := indentSize * level
	usableScreenSize := screenWidth - currentIndentSize

	if commentWidth == 0 {
		return max(usableScreenSize, 40)
	}
	if usableScreenSize < commentWidth {
		return usableScreenSize
	}

	return commentWidth
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
func labelAuthor(author, originalPoster, parentPoster string) string {
	authorInBold := bold(author)

	switch author {
	case "dang":
		return authorInBold + green(" mod")
	case "sctb":
		return authorInBold + green(" mod")
	case originalPoster:
		return authorInBold + red(" OP")
	case parentPoster:
		return authorInBold + purple(" PP")
	default:
		return authorInBold
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
	for i := 0; i < indentSize*level; i++ {
		indentation = " " + indentation
	}
	return indentation
}

func parseComment(comment string) (string, []string) {
	comment = replaceCharacters(comment)
	comment = replaceHTML(comment)
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
	return input
}

func replaceHTML(input string) string {
	input = strings.Replace(input, "<p>", "", 1)

	input = strings.ReplaceAll(input, "<p>", DoubleNewLine)
	input = strings.ReplaceAll(input, "<i>", Italic)
	input = strings.ReplaceAll(input, "</i>", Normal)
	input = strings.ReplaceAll(input, "</a>", "")
	input = strings.ReplaceAll(input, "<pre><code>", Dimmed)
	input = strings.ReplaceAll(input, "</code></pre>", Normal)
	return input
}

func extractURLs(input string) []string {
	expForFirstTag := regexp.MustCompile(`<a href=".*?" rel="nofollow">`)
	URLs := expForFirstTag.FindAllString(input, 10)

	for i, _ := range URLs {
		URLs[i] = strings.ReplaceAll(URLs[i], `<a href="`, "")
		URLs[i] = strings.ReplaceAll(URLs[i], `" rel="nofollow">`, "")
	}

	return URLs
}

func trimURLs(comment string) string {
	expression := regexp.MustCompile(`<a href=".*?" rel="nofollow">`)
	return expression.ReplaceAllString(comment, "")
}
