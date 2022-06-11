package message

import "clx/item"

type EditorFinishedMsg struct {
	Err error
}

type EnteringCommentSection struct {
	Id           int
	CommentCount int
}

type EnteringReaderMode struct {
	Url    string
	Title  string
	Domain string
}

type EnterHelpScreen struct{}

type StatusMessageTimeout struct{}

type FetchingFinished struct{}

type ChangeCategory struct {
	Category int
	Cursor   int
}

type CategoryFetchingFinished struct {
	Category int
	Cursor   int
}

type AddToFavorites struct {
	Item *item.Item
}
