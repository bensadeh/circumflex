package main

import (
	"regexp"
	"strconv"
	"strings"

	term "github.com/MichaelMure/go-term-text"
	"github.com/eidolon/wordwrap"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

type Comments struct {
	Author        string      `json:"user"`
	Title         string      `json:"title"`
	Comment       string      `json:"content"`
	CommentsCount int         `json:"comments_count"`
	Time          string      `json:"time_ago"`
	Points        int         `json:"points"`
	URL           string      `json:"url"`
	Domain        string      `json:"domain"`
	Replies       []*Comments `json:"comments"`
}

func appendCommentsHeader(comment Comments, commentTree *string) {
	headline := "\033[1m" + comment.Title + "\033[0m" + "\033[2m" + "  (" + comment.Domain + ")" + "\033[0m" + "\n"
	*commentTree += headline
	*commentTree += strconv.Itoa(comment.Points) + " points by " + "\033[1m" + comment.Author + "\033[0m" + " " + comment.Time + " | " + strconv.Itoa(comment.CommentsCount) + " comments" + "\n"
	for i := 0; i < term.Len(headline); i++ {
		*commentTree += "-"
	}
	*commentTree += "\n\n"
}

func prettyPrintComments(c Comments, commentTree *string, indentlevel int, op string) string {
	x, _ := terminal.Width()
	rightPadding := 3
	comment := parseComment(c.Comment)
	wrapper := wordwrap.Wrapper(int(x)-indentlevel-rightPadding, false)
	markedAuthor := markOPAndMods(c.Author, op)

	fullComment := ""
	commentLines := strings.Split(comment, "\n")
	for _, line := range commentLines {
		wrapped := wrapper(line)
		wrappedAndIndentedComment := wordwrap.Indent(wrapped, getIndentBlock(indentlevel), true)
		fullComment += wrappedAndIndentedComment + "\n" + "\n"
	}

	wrappedAndIndentedAuthor := wordwrap.Indent(markedAuthor, getIndentBlock(indentlevel), true)
	wrappedAndIndentedComment := "\033[1m" + wrappedAndIndentedAuthor + "\033[0m " + getRightAlignedTimeAgo(markedAuthor, c.Time, indentlevel)
	wrappedAndIndentedComment += fullComment

	*commentTree = *commentTree + wrappedAndIndentedComment
	for _, s := range c.Replies {
		prettyPrintComments(*s, commentTree, indentlevel+5, op)
	}
	return *commentTree
}

func getRightAlignedTimeAgo(author string, timeAgo string, indentLevel int) string {
	screenWidth, _ := terminal.Width()
	authorLength := term.Len(author)
	timeAgoLength := term.Len(timeAgo)
	paddingBetweenAuthorAndTime := ""
	padding := 6

	numberOfSpaces := int(screenWidth) - authorLength - timeAgoLength - padding - indentLevel

	for i := 0; i < numberOfSpaces; i++ {
		paddingBetweenAuthorAndTime += " "
	}

	return paddingBetweenAuthorAndTime + "\033[2m" + timeAgo + "\033[0m\n"

}

func markOPAndMods(author, op string) string {
	markedAuthor := author
	if author == "dang" || author == "sctb" {
		markedAuthor = author + "\033[32;m" + " mod" + "\033[0m"
	}
	if author == op {
		markedAuthor = markedAuthor + "\033[31;m" + " OP" + "\033[0m"
	}
	return markedAuthor
}

func getIndentBlock(level int) string {
	indentation := " "
	for i := 1; i < level; i++ {
		indentation = indentation + " "
	}
	return indentation
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

	input = strings.ReplaceAll(input, "<p>", "\n")
	input = strings.ReplaceAll(input, "<i>", "\033[3m")
	input = strings.ReplaceAll(input, "</i>", "\033[0m")
	input = strings.ReplaceAll(input, "<pre><code>", "\033[2m")
	input = strings.ReplaceAll(input, "</code></pre>", "\033[0m")
	return input
}

func handleHrefTag(input string) string {
	var expForFirstTag = regexp.MustCompile(`<a href="`)
	replacedInput := expForFirstTag.ReplaceAllString(input, "\033]8;;")

	var expForSecondTag = regexp.MustCompile(`" rel="nofollow">`)
	replacedInput = expForSecondTag.ReplaceAllString(replacedInput, "\a")

	var expForThirdTag = regexp.MustCompile(`<\/a>`)
	replacedInput = expForThirdTag.ReplaceAllString(replacedInput, "\033]8;;\a")

	return replacedInput
}
