package message

type EditorFinishedMsg struct{ Err error }
type EnteringCommentSectionMsg struct{ Id int }
type StatusMessageTimeoutMsg struct{}
type FetchingFinished struct{}
