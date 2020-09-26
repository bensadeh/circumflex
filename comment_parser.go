package main

import (
	"strconv"
	"strings"

	"clx/wordwrap"

	term "github.com/MichaelMure/go-term-text"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

// Comments represent the JSON structure as
// retreived from cheeaun's unoffical HN API
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

func printCommentTree(comments Comments, indentSize int, commmentWidth int) string {
	header := getHeader(comments)
	originalPoster := comments.Author
	commentTree := ""
	for _, reply := range comments.Replies {
		commentTree += prettyPrintComments(*reply, 0, indentSize, commmentWidth, originalPoster)
	}
	return header + commentTree
}

func getHeader(c Comments) string {
	headline := c.Title + getDomainText(c.Domain, c.URL, c.ID) + NewLine
	headlineWithoutHyperlink := c.Title + getDomainTextWithoutHyperlink(c.Domain, c.URL, c.ID) + NewLine
	headlineWithoutHyperlinkLength := term.Len(headlineWithoutHyperlink)
	infoLine := dimmed(strconv.Itoa(c.Points)+" points by "+c.Author+" "+c.Time+" • "+strconv.Itoa(c.CommentsCount)+" comments") + NewLine

	infoLineLength := term.Len(infoLine)
	longestLine := max(headlineWithoutHyperlinkLength, infoLineLength)

	header := headline + infoLine
	header += parseRootComment(c.Comment, longestLine)

	for i := 0; i < longestLine; i++ {
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

func getDomainTextWithoutHyperlink(domain string, URL string, id int) string {
	if domain != "" {
		return " " + paren(domain)
	}
	linkText := "item?id=" + strconv.Itoa(id)
	return " " + paren(linkText)
}

func parseRootComment(comment string, lineLength int) string {
	if comment == "" {
		return ""
	}

	parsedComment := parseComment(comment)

	commentLines := strings.Split(parsedComment, "<p>")
	lastParagraph := len(commentLines) - 1
	firstParagraph := 0
	fullComment := ""
	for i, line := range commentLines {
		wrapped := wordwrap.WrapString(line, uint(lineLength))
		wrappedAndIndentedComment := wordwrap.Indent(wrapped, getIndentBlock(0, 0), true)
		if i == firstParagraph {
			fullComment = NewLine
		}
		if i == lastParagraph {
			fullComment += wrappedAndIndentedComment + NewLine
		} else {
			fullComment += wrappedAndIndentedComment + DoubleNewLine
		}
	}
	return fullComment
}

func prettyPrintComments(c Comments, level int, indentSize int, commmentWidth int, op string) string {
	comment := parseComment(c.Comment)
	limit := getCommentWidth(level, indentSize, commmentWidth)
	markedAuthor := markOPAndMods(c.Author, op)

	paragraphs := strings.Split(comment, "<p>")
	lastParagraph := len(paragraphs) - 1
	fullComment := ""
	for i, paragraph := range paragraphs {
		wrappedParagraph := wordwrap.WrapString(paragraph, uint(limit))
		wrappedAndIndentedParagraph := wordwrap.Indent(wrappedParagraph, getIndentBlock(level, indentSize), true)

		if i == lastParagraph {
			fullComment += wrappedAndIndentedParagraph + DoubleNewLine
			break
		}

		barOnEmptyLine := wordwrap.Indent("", getIndentBlock(level, indentSize), true)
		fullComment += wrappedAndIndentedParagraph + NewLine + barOnEmptyLine + NewLine
	}

	author := wordwrap.Indent(markedAuthor, getIndentBlockWithoutBar(level, indentSize), true)
	authorAndTimeStamp := author + " " + dimmed(c.Time) + getTopLevelCommentAnchor(level) + NewLine
	fullCommentWithAuthor := authorAndTimeStamp + fullComment

	for _, s := range c.Replies {
		fullCommentWithAuthor += prettyPrintComments(*s, level+1, indentSize, commmentWidth, op)
	}
	return fullCommentWithAuthor
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func getTopLevelCommentAnchor(level int) string {
	if level == 0 {
		return " ::"
	}
	return ""
}

func getCommentWidth(level int, indentSize int, commentWidth int) int {
	x, _ := terminal.Width()
	screenWidth := int(x)
	// hack: the wrapper is sometimes off by 1, so we pad
	// the wrapper to end the line slightly earlier
	padding := 1
	actualIndentSize := indentSize * level
	usableScreenSize := screenWidth - actualIndentSize - padding

	if commentWidth == 0 {
		return max(usableScreenSize, 40)
	}
	if usableScreenSize < commentWidth {
		return usableScreenSize
	}

	return commentWidth
}

func markOPAndMods(author, op string) string {
	markedAuthor := bold(author)
	if author == "dang" || author == "sctb" {
		markedAuthor = markedAuthor + green(" mod")
	}
	if author == op {
		markedAuthor = markedAuthor + red(" OP")
	}
	return markedAuthor
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

func parseComment(comment string) string {
	comment = replaceHTML(comment)
	comment = replaceCharacters(comment)
	return comment
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

	input = strings.ReplaceAll(input, "<i>", Italic)
	input = strings.ReplaceAll(input, "</i>", Normal)
	input = strings.ReplaceAll(input, "<pre><code>", Dimmed)
	input = strings.ReplaceAll(input, "</code></pre>", Normal)
	input = strings.ReplaceAll(input, `<a href="`, Link1)
	input = strings.ReplaceAll(input, `" rel="nofollow">`, Link2)
	input = strings.ReplaceAll(input, `</a>`, Link3)
	return input
}
