package main

import (
	"fmt"
	// "circumflex/client"
	"circumflex/cmd"
	"encoding/json"
	"flag"

	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/TylerBrock/colorjson"
	// "github.com/gdamore/tcell"
	"github.com/gocolly/colly"

	"gitlab.com/tslocum/cview"
	// "github.com/rivo/tview"
	"github.com/eidolon/wordwrap"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

func main() {
	cmd.Execute()

	// client := client.NewHNClient()
	// pp, err := client.GetTopStories(30)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// // for _, v := range *pp {
	// // 	fmt.Println(v.Title)
	// // }

	// app := cview.NewApplication()
	// list := cview.NewList()

	// list.SetBackgroundTransparent(false)
	// list.SetBackgroundColor(tcell.ColorDefault)
	// list.SetMainTextColor(tcell.ColorDefault)
	// list.SetSecondaryTextColor(tcell.ColorGray)
	// list.ShowSecondaryText(false)

	// reset := func() {
	// 	list.Clear()
	// 	for _, s := range *pp {
	// 		list.AddItem(s.Title, s.Author, 0, nil)
	// 	}
	// }

	// reset()
	// if err := app.SetRoot(list, true).EnableMouse(true).Run(); err != nil {
	// 	panic(err)
	// }

	comments()
}

type comment struct {
	Author  string `selector:"a.hnuser"`
	URL     string `selector:".age a[href]" attr:"href"`
	Comment string `selector:".comment"`
	Replies []*comment
	depth   int
}

func comments() {
	var itemID string
	flag.StringVar(&itemID, "id", "24089281", "hackernews post id")
	flag.Parse()

	if itemID == "" {
		log.Println("Hackernews post id required")
		os.Exit(1)
	}

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

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	// Dump json to the standard output
	// enc.Encode(comments)

	f := colorjson.NewFormatter()
	f.Indent = 2
	f.RawStrings = false

	s, _ := f.Marshal(comments)
	fmt.Println(string(s))

	// Pager logic
	// pager := os.ExpandEnv("$PAGER")

	// Could read $PAGER rather than hardcoding the path.
	cmd := exec.Command("/usr/bin/less")

	// stringComments := ""
	// for _, s := range comments {
	// 	stringComments = stringComments + s.Author + ": " + s.Comment + "\n"
	// 	for _, t := range s.Replies {
	// 		wrapper := wordwrap.Wrapper(20, false)
	// 		wrapped := wrapper(t.Comment)
	// 		indented := wordwrap.Indent(wrapped, "            ", true)

	// 		stringComments = stringComments + t.Author + ": " + indented
	// 	}
	// }

	commentTree := ""
	for _, s := range comments {
		commentTree = prettyPrintComments(*s, &commentTree, 0)

	}

	// Feed it with the string you want to display.
	// cmd.Stdin = strings.NewReader(stringComments)
	cmd.Stdin = strings.NewReader(commentTree)

	// This is crucial - otherwise it will write to a null device.
	cmd.Stdout = os.Stdout

	// Fork off a process and wait for it to terminate.
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	tree := cview.NewTreeView()

	root := cview.NewTreeNode("Root").
		AddChild(cview.NewTreeNode("First Child").
			AddChild(cview.NewTreeNode("GrandChild")).
			AddChild(cview.NewTreeNode("GrandChild"))).
		AddChild(cview.NewTreeNode("Second Child")).
		AddChild(cview.NewTreeNode("GrandChild")).
		AddChild(cview.NewTreeNode("GrandChild"))

	tree.SetRoot(root).SetCurrentNode(root)

	cview.NewApplication().SetRoot(tree, true).Run()

}

func prettyPrintComments(c comment, commentTree *string, indentlevel int) string {
	x, _ := terminal.Width()
	wrapper := wordwrap.Wrapper(int(x) - indentlevel - 1, false)
	wrapped := wrapper(c.Author + ": " + c.Comment)
	wrappedAndIndentedComment := wordwrap.Indent(wrapped, getindent(indentlevel), true)
	wrappedAndIndentedComment = "\n" + wrappedAndIndentedComment + "\n"

	*commentTree = *commentTree + wrappedAndIndentedComment
	for _, s := range c.Replies {
		prettyPrintComments(*s, commentTree, indentlevel + 10)
	}
	return *commentTree
}

func getindent(level int) string {
	indentation := " "
	for i := 1; i < level; i++ {
		indentation = indentation + " "
	}
	return indentation
}
