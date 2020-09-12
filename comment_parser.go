package main

import (
	"regexp"
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

func appendCommentsHeader(c Comments, commentTree *string) {
	headline := bold(c.Title) + getDomainText(c.Domain, c.URL, c.ID) + NewLine
	headlineWithoutHyperlink := bold(c.Title) + getDomainTextWithoutHyperlink(c.Domain, c.URL, c.ID) + NewLine
	infoLine := strconv.Itoa(c.Points) + " points by " + bold(c.Author) + " " + c.Time + " • " + strconv.Itoa(c.CommentsCount) + " comments" + NewLine
	*commentTree += headline + infoLine
	*commentTree += parseRootComment(c.Comment)

	headlineWithoutHyperlinkLength := term.Len(headlineWithoutHyperlink)
	for i := 0; i < headlineWithoutHyperlinkLength; i++ {
		*commentTree += "-"
	}

	*commentTree += DoubleNewLine
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

func parseRootComment(comment string) string {
	if comment == "" {
		return ""
	}

	x, _ := terminal.Width()
	wrapper := wordwrap.Wrapper(int(x), false)
	parsedComment := parseComment(comment)

	commentLines := strings.Split(parsedComment, NewLine)
	lastParagraph := len(commentLines) - 1
	firstParagraph := 0
	fullComment := ""
	for i, line := range commentLines {
		wrapped := wrapper(line)
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

func prettyPrintComments(c Comments, commentTree *string, level int, indentSize int, commmentWidth int, op string) string {
	comment := parseComment(c.Comment)
	limit := getCommentWidth(level, indentSize, commmentWidth)
	wrapper := wordwrap.Wrapper(limit, false)
	markedAuthor := markOPAndMods(c.Author, op)

	fullComment := ""
	paragraphs := strings.Split(comment, NewLine)
	lastParagraph := len(paragraphs) - 1
	for i, paragraph := range paragraphs {
		wrapped := wrapper(paragraph)
		wrappedAndIndentedComment := wordwrap.Indent(wrapped, getIndentBlock(level, indentSize), true)
		barOnEmptyLine := wordwrap.Indent("", getIndentBlock(level, indentSize), true)

		if i == lastParagraph {
			fullComment += wrappedAndIndentedComment + DoubleNewLine
			break
		}
		fullComment += wrappedAndIndentedComment + NewLine + barOnEmptyLine + NewLine
	}

	wrappedAndIndentedAuthor := wordwrap.Indent(markedAuthor, getIndentBlockWithoutBar(level, indentSize), true)
	wrappedAndIndentedComment := wrappedAndIndentedAuthor + " " + dimmed(c.Time) + NewLine
	wrappedAndIndentedComment += fullComment

	*commentTree = *commentTree + wrappedAndIndentedComment
	for _, s := range c.Replies {
		prettyPrintComments(*s, commentTree, level+1, indentSize, commmentWidth, op)
	}
	return *commentTree
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
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
	indentation := getColoredIndentBlock(level) + "▎" + Normal
	for i := 0; i < indentSize*level; i++ {
		indentation = " " + indentation
	}
	return indentation
}

func getColoredIndentBlock(level int) string {
	switch level {
	case 1:
		return Red
	case 2:
		return Yellow
	case 3:
		return Green
	case 4:
		return Blue
	case 5:
		return Teal
	case 6:
		return Purple
	case 7:
		return Red
	case 8:
		return Yellow
	case 9:
		return Green
	case 10:
		return Blue
	case 11:
		return Teal
	case 12:
		return Purple
	default:
		return Normal
	}
}

func parseComment(comment string) string {
	fixedHTML := replaceHTML(comment)
	fixedHTMLAndCharacters := replaceCharacters(fixedHTML)
	fixedHTMLAndCharactersAndHrefs := handleHrefTag(fixedHTMLAndCharacters)
	return fixedHTMLAndCharactersAndHrefs
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

	input = strings.ReplaceAll(input, "<p>", NewLine)
	input = strings.ReplaceAll(input, "<i>", Italic)
	input = strings.ReplaceAll(input, "</i>", Normal)
	input = strings.ReplaceAll(input, "<pre><code>", Dimmed)
	input = strings.ReplaceAll(input, "</code></pre>", Normal)
	return input
}

func handleHrefTag(input string) string {
	var expForFirstTag = regexp.MustCompile(`<a href="`)
	replacedInput := expForFirstTag.ReplaceAllString(input, Link1)

	var expForSecondTag = regexp.MustCompile(`" rel="nofollow">`)
	replacedInput = expForSecondTag.ReplaceAllString(replacedInput, Link2)

	var expForThirdTag = regexp.MustCompile(`<\/a>`)
	replacedInput = expForThirdTag.ReplaceAllString(replacedInput, Link3)

	return replacedInput
}
