package hn

import (
	"context"
)

// Story represents a Hacker News story as returned by FetchItems/FetchItem.
// This is the list-view representation: no comment tree, no self-text.
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

type CommentTree struct {
	ID            int
	Title         string
	Points        int
	Author        string
	Time          int64
	TimeAgo       string
	URL           string
	Domain        string
	Content       string
	CommentsCount int
	Comments      []*CommentNode
}

type CommentNode struct {
	ID       int
	Author   string
	Time     int64
	TimeAgo  string
	Content  string
	Children []*CommentNode
}

type Service interface {
	FetchItems(ctx context.Context, itemsToFetch int, category string) ([]*Story, error)
	FetchItem(ctx context.Context, id int) (*Story, error)
	FetchComments(ctx context.Context, id int, onProgress func(fetched, total int)) (*CommentTree, error)
}
