package message

import (
	"clx/article"
	"clx/categories"
	"clx/comment"
	"clx/item"
	"time"
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
