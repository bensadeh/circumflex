package history

type History interface {
	Contains(id int) bool
	GetLastVisited(id int) int64
	GetLastCommentCount(id int) int
	ClearAndWriteToDisk()
	AddToHistoryAndWriteToDisk(id int, commentsOnLastVisit int)
}
