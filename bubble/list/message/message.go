package message

type EditorFinishedMsg struct{ Err error }
type EnteringCommentSection struct{ Id int }
type StatusMessageTimeout struct{}
type FetchingFinished struct{}
