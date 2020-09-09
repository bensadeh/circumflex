package main

import (
	"github.com/eidolon/wordwrap"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func prettyPrintComments(c Comments, commentTree *string, indentlevel int) string {
	x, _ := terminal.Width()
	rightPadding := 3
	wrapper := wordwrap.Wrapper(int(x)-indentlevel-rightPadding, false)
	wrapped := wrapper(c.Comment)
	wrappedAndIndentedComment := wordwrap.Indent(wrapped, getIndentBlock(indentlevel), true)
	wrappedAndIndentedAuthor := wordwrap.Indent(c.Author, getIndentBlock(indentlevel), true)

	wrappedAndIndentedComment = "\033[1m" + wrappedAndIndentedAuthor + "\033[21m" + "\n" + wrappedAndIndentedComment + "\n" + "\n"

	*commentTree = *commentTree + wrappedAndIndentedComment
	for _, s := range c.Replies {
		prettyPrintComments(*s, commentTree, indentlevel+5)
	}
	return *commentTree
}

func getIndentBlock(level int) string {
	indentation := " "
	for i := 1; i < level; i++ {
		indentation = indentation + " "
	}
	return indentation
}
