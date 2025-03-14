package message

import (
	"clx/item"
	"time"
)

type EditorFinishedMsg struct {
	Err error
}

type EnteringCommentSection struct {
	Id           int
	CommentCount int
}

type OpeningCommentsInBrowser struct {
	Id           int
	CommentCount int
}

type OpeningLink struct {
	Id           int
	CommentCount int
}

type EnteringReaderMode struct {
	Url          string
	Title        string
	Domain       string
	Id           int
	CommentCount int
}

type ShowStatusMessage struct {
	Message  string
	Duration time.Duration
}

type StatusMessageTimeout struct{}

type FetchingFinished struct {
	Message string
}

type FetchAndChangeToCategory struct {
	Index    int
	Category int
	Cursor   int
}

type Refresh struct {
	CurrentCategory int
	CurrentIndex    int
}

type CategoryFetchingFinished struct {
	Index   int
	Cursor  int
	Message string
}

type AddToFavorites struct {
	Item *item.Item
}

type DownloadFile struct {
	Url string
}
