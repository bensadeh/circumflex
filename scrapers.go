package main

import (
	"strconv"
	"strings"

	"github.com/eidolon/wordwrap"
	"github.com/gocolly/colly"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

type comment struct {
	Author  string `selector:"a.hnuser"`
	URL     string `selector:".age a[href]" attr:"href"`
	Comment string `selector:".comment"`
	Replies []*comment
	depth   int
}

func scrapeComments(itemID string) string {
	comments := make([]*comment, 0)

	// Instantiate default collector
	c := colly.NewCollector()

	// Extract comment
	c.OnHTML(".comment-tree tr.athing", func(e *colly.HTMLElement) {
		width, err := strconv.Atoi(e.ChildAttr("td.ind img", "width"))
		if err != nil {
			return
		}
		// hackernews uses 40px spacers to indent comment replies,
		// so we have to divide the width with it to get the depth
		// of the comment
		depth := width / 40
		c := &comment{
			Replies: make([]*comment, 0),
			depth:   depth,
		}
		e.Unmarshal(c)
		c.Comment = strings.TrimSpace(c.Comment[:len(c.Comment)-5])
		if depth == 0 {
			comments = append(comments, c)
			return
		}
		parent := comments[len(comments)-1]
		// append comment to its parent
		for i := 0; i < depth-1; i++ {
			parent = parent.Replies[len(parent.Replies)-1]
		}
		parent.Replies = append(parent.Replies, c)
	})

	c.Visit("https://news.ycombinator.com/item?id=" + itemID)

	commentTree := ""
	for _, s := range comments {
		commentTree = prettyPrintComments(*s, &commentTree, 0)

	}

	return commentTree
}

func prettyPrintComments(c comment, commentTree *string, indentlevel int) string {
	x, _ := terminal.Width()
	wrapper := wordwrap.Wrapper(int(x)-indentlevel-1, false)
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
