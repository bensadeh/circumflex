package main

import (
	"regexp"
	"strings"

	"github.com/eidolon/wordwrap"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func appendCommentsHeader(comment Comments, commentTree *string) {
	*commentTree += "\033[7m" + comment.Title + "\033[0m" + "\n" + "\n"
}

func prettyPrintComments(c Comments, commentTree *string, indentlevel int, op string) string {
	x, _ := terminal.Width()
	rightPadding := 3
	comment := parseComment(c.Comment)
	wrapper := wordwrap.Wrapper(int(x)-indentlevel-rightPadding, false)
	markedAuthor := markOP(c.Author, op)

	fullComment := ""
	commentLines := strings.Split(comment, "\n")
	for _, line := range commentLines {
		wrapped := wrapper(line)
		wrappedAndIndentedComment := wordwrap.Indent(wrapped, getIndentBlock(indentlevel), true)
		fullComment += wrappedAndIndentedComment + "\n" + "\n"
	}

	wrappedAndIndentedAuthor := wordwrap.Indent(markedAuthor, getIndentBlock(indentlevel), true)
	wrappedAndIndentedComment := "\033[1m" + wrappedAndIndentedAuthor + "\033[0m" + "\n" + fullComment

	*commentTree = *commentTree + wrappedAndIndentedComment
	for _, s := range c.Replies {
		prettyPrintComments(*s, commentTree, indentlevel+5, op)
	}
	return *commentTree
}

func markOP(author, op string) string {
	if author == op {
		return author + "\033[31;m" + " OP" + "\033[0m"
	}
	return author
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
	replacedInput := expForFirstTag.ReplaceAllString(input, "\033[4m")

	var validID = regexp.MustCompile(`" rel="nofollow">(.*?)<\/a>`)
	return validID.ReplaceAllString(replacedInput, "\033[0m")
}
