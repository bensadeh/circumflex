package history

import "time"

type NonPersistent struct{}

func (NonPersistent) Contains(_ int) bool {
	return false
}

func (NonPersistent) CommentsLastVisited(_ int) int64 {
	return time.Now().Unix()
}

func (NonPersistent) Clear() error { return nil }

func (NonPersistent) MarkRead(_ int, _ int) error { return nil }

func (NonPersistent) MarkArticleRead(_ int) error { return nil }

func (NonPersistent) MarkUnread(_ int) error { return nil }
