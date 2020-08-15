package main

import (
	// "circumflex/client"
	"circumflex/cmd"
	"encoding/json"
	"flag"

	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"

	// "github.com/gdamore/tcell"
	"github.com/gocolly/colly"
	// "gitlab.com/tslocum/cview"
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
	enc.Encode(comments)

//########################

	// Observe how the b's and the d's, despite appearing in the
	// second cell of each line, belong to different columns.
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, '.', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Fprintln(w, "a\tb\tc")
	fmt.Fprintln(w, "aa\tbb\tcc")
	fmt.Fprintln(w, "aaa\t") // trailing tab
	fmt.Fprintln(w, "aaaa\tdddd\teeee")
	w.Flush()

	for _, s := range comments {
		fmt.Fprintln(w, s.Author + ": " + s.Comment)
		for _, t := range s.Replies {
			fmt.Fprintln(w, "\t" + t.Author + ": " + t.Comment)

		}
	}

//########################



	// Pager logic
	// pager := os.ExpandEnv("$PAGER")

	// Could read $PAGER rather than hardcoding the path.
	cmd := exec.Command("/usr/bin/less")

	// stringComments := ""
	// for _, s := range comments {
	// 	stringComments = stringComments + s.Author + ": " + s.Comment +  "\n"
	// 	for _, t := range s.Replies {
	// 		stringComments = stringComments + t.Author + ": " + t.Comment
	// 	}
	// }

	// Feed it with the string you want to display.
	// cmd.Stdin = strings.NewReader(stringComments)

	// This is crucial - otherwise it will write to a null device.
	cmd.Stdout = os.Stdout

	// Fork off a process and wait for it to terminate.
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

}
