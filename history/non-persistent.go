package history

import "time"

type NonPersistent struct{}

func (NonPersistent) Contains(_ int) bool {
	return false
}

func (NonPersistent) GetLastVisited(_ int) int64 {
	return time.Now().Unix()
}

func (NonPersistent) GetLastCommentCount(_ int) int {
	return 0
}

func (NonPersistent) ClearAndWriteToDisk() {}

func (NonPersistent) MarkAsReadAndWriteToDisk(_ int, _ int) {}
