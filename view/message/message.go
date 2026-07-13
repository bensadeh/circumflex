package message

import (
	"context"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/browser"
	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/comment"
	"github.com/bensadeh/circumflex/hn"

	tea "charm.land/bubbletea/v2"
)

// DetailQuit closes the open detail view — comments or reader — and returns
// to the story list.
type DetailQuit struct{}

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

type StatusMessageTimeout struct {
	Generation int
}

// StoriesReady delivers a category's stories: a tab switch, a refresh, or
// the startup fetch (including the fabricated result the local favorites
// category is served through). Index and Cursor are the category selection
// and list cursor to land on when the stories apply.
type StoriesReady struct {
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

// OpenReaderLink asks the app to open a link followed from inside a
// reader-mode article; the linked page replaces the article in the pane.
type OpenReaderLink struct {
	URL string
}

func OpenReaderLinkCmd(url string) tea.Cmd {
	return func() tea.Msg { return OpenReaderLink{URL: url} }
}

// LinkArticleReady delivers a page fetched through OpenReaderLink. Kept
// apart from ArticleReady because failure handling differs: the open article
// stays on screen, nothing rolls back, and no history is marked.
type LinkArticleReady struct {
	Parsed  *article.Parsed
	Title   string
	URL     string
	Err     error
	FetchID uint64
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

// ErrorViewQuit closes the error view a failed story load left in the
// detail pane.
type ErrorViewQuit struct{}

// ErrorProgressTimeout settles the terminal progress indicator a few seconds
// after a failed story load left it in its error state, like a status
// message expiring. FetchID guards a newer fetch's indicator from being
// cleared by a stale timeout.
type ErrorProgressTimeout struct {
	FetchID uint64
}

// OpenAdjacentStory asks the front page to open the next (+1) or previous
// (-1) story in the same view the request came from: comment section or
// reader mode. A direction of 0 re-opens the selected story itself — the
// reader's stateless step back from a followed link.
type OpenAdjacentStory struct {
	Direction int
}

func OpenAdjacentStoryCmd(direction int) tea.Cmd {
	return func() tea.Msg { return OpenAdjacentStory{Direction: direction} }
}

type TimeRefreshTick struct{}

type MemorialStatusReady struct {
	Active bool
	Err    error
}
