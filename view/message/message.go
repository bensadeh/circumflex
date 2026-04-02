package message

import (
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/item"
)

type ReaderViewQuitMsg struct{}

type EnteringCommentSection struct {
	ID           int
	CommentCount int
}

type BrowserOpenFailed struct {
	Err error
}

type EnteringReaderMode struct {
	URL          string
	Title        string
	Domain       string
	ID           int
	CommentCount int
	Author       string
	TimeAgo      string
	Points       int
}

type ShowStatusMessage struct {
	Message  string
	Duration time.Duration
}

type StatusMessageTimeout struct {
	Generation int
}

type FetchingFinished struct {
	Stories  []*item.Story
	Category categories.Category
	Err      error
}

type FetchAndChangeToCategory struct {
	Index    int
	Category categories.Category
	Cursor   int
}

type Refresh struct {
	CurrentCategory categories.Category
	CurrentIndex    int
}

type CategoryFetchingFinished struct {
	Stories  []*item.Story
	Category categories.Category
	Index    int
	Cursor   int
	Err      error
	FetchID  uint64
}

type AddToFavorites struct {
	Item *item.Story
}

type ArticleReady struct {
	Parsed  *article.Parsed
	Content string
	Title   string
	URL     string
	Author  string
	TimeAgo string
	ID      int
	Points  int
	Err     error
	FetchID uint64
}

type CommentTreeDataReady struct {
	Thread       *comment.Thread
	LastVisited  int64
	UpdatedStory *item.Story // stays in item domain for favorites sync
	Err          error
	FetchID      uint64
}

type CommentViewQuitMsg struct{}

type TimeRefreshTick struct{}
