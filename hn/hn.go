package hn

import (
	"context"
	"strconv"
)

// ItemURL returns the Hacker News page for an item (story or comment thread).
func ItemURL(id int) string {
	return "https://news.ycombinator.com/item?id=" + strconv.Itoa(id)
}

// Story is a Hacker News story's metadata as returned by
// FetchItems/FetchItem: no comment tree, no self-text. Story-shaped types
// (CommentTree, comment.Thread) embed it, so a field added here flows to
// them without hand copying.
type Story struct {
	ID            int
	Title         string
	Points        int
	Author        string
	Time          int64
	URL           string
	Domain        string
	CommentsCount int
}

// CommentTree is a story plus its self-text and fetched comment tree.
type CommentTree struct {
	Story

	Content  string
	Comments []*CommentNode
}

type CommentNode struct {
	ID       int
	Author   string
	Time     int64
	Content  string
	Children []*CommentNode
}

type Service interface {
	FetchItems(ctx context.Context, itemsToFetch int, category string) ([]*Story, error)
	FetchItem(ctx context.Context, id int) (*Story, error)
	FetchComments(ctx context.Context, id int, onProgress func(fetched, total int)) (*CommentTree, error)
}
