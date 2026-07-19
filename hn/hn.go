package hn

import (
	"context"
	nurl "net/url"
	"strconv"
	"strings"
	"time"
)

// ItemURL returns the Hacker News page for an item (story or comment thread).
func ItemURL(id int) string {
	return "https://news.ycombinator.com/item?id=" + strconv.Itoa(id)
}

// ParseItemURL is ItemURL's reverse: the item ID a Hacker News discussion
// link points at. Only item pages match — other HN pages (user profiles,
// front pages) carry no thread to open.
func ParseItemURL(rawURL string) (int, bool) {
	u, err := nurl.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return 0, false
	}

	if strings.TrimPrefix(u.Hostname(), "www.") != "news.ycombinator.com" || u.Path != "/item" {
		return 0, false
	}

	id, err := strconv.Atoi(u.Query().Get("id"))
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
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

// SearchRequest is one story search: the query plus the filters mirroring
// hn.algolia.com's dropdowns — how to rank and how far back.
type SearchRequest struct {
	Query        string
	SortByDate   bool          // false ranks by relevance/popularity
	MaxAge       time.Duration // 0 means all time
	ItemsToFetch int
}

type Service interface {
	FetchItems(ctx context.Context, itemsToFetch int, category string) ([]*Story, error)
	FetchItem(ctx context.Context, id int) (*Story, error)
	FetchComments(ctx context.Context, id int, onProgress func(fetched, total int)) (*CommentTree, error)
	SearchItems(ctx context.Context, req SearchRequest) ([]*Story, error)
}
