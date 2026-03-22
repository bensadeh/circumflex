package history

import "time"

type NonPersistent struct{}

func (NonPersistent) Contains(_ int) bool {
	return false
}

func (NonPersistent) CommentsLastVisited(_ int) int64 {
	return time.Now().Unix()
}

func (NonPersistent) ClearAndWriteToDisk() error { return nil }

func (NonPersistent) MarkAsReadAndWriteToDisk(_ int, _ int) error { return nil }

func (NonPersistent) MarkArticleAsReadAndWriteToDisk(_ int) error { return nil }

func (NonPersistent) MarkAsUnreadAndWriteToDisk(_ int) error { return nil }
