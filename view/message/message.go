package message

import (
	"context"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/browser"
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/hn"

	tea "charm.land/bubbletea/v2"
)

type ReaderViewQuit struct{}

type EnteringCommentSection struct {
	ID           int
	CommentCount int
}

type BrowserOpenFailed struct {
	Err error
}

func OpenInBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		if err := browser.Open(context.Background(), url); err != nil {
			return BrowserOpenFailed{Err: err}
		}

		return nil
	}
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
	Stories  []*hn.Story
	Category categories.Category
	Err      error
	FetchID  uint64
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
	Stories  []*hn.Story
	Category categories.Category
	Index    int
	Cursor   int
	Err      error
	FetchID  uint64
}

type AddToFavorites struct {
	Item *hn.Story
}

type ArticleReady struct {
	Parsed         *article.Parsed
	Title          string
	URL            string
	Author         string
	TimeAgo        string
	ID             int
	Points         int
	Err            error
	FetchID        uint64
	HistoryWarning error // non-nil if marking as read failed
}

type CommentTreeDataReady struct {
	Thread         *comment.Thread
	LastVisited    int64
	UpdatedStory   *hn.Story // non-nil when viewing a favorite, for syncing metadata
	Err            error
	FetchID        uint64
	HistoryWarning error // non-nil if marking as read failed
}

type CommentViewQuit struct{}

type TimeRefreshTick struct{}

type MemorialStatusReady struct {
	Active bool
	Err    error
}
