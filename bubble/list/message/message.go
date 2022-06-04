package message

type EditorFinishedMsg struct {
	Err error
}

type EnteringCommentSection struct {
	Id           int
	CommentCount int
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
