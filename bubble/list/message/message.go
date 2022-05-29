package message

type EditorFinishedMsg struct {
	Err error
}

type EnteringCommentSection struct {
	Id           int
	CommentCount int
}

type StatusMessageTimeout struct{}

type FetchingFinished struct{}

type ChangeCategory struct {
	Category              int
	ItemCurrentlySelected int
}

type CategoryFetchingFinished struct {
	Category              int
	ItemCurrentlySelected int
}
