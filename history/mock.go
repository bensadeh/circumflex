package history

import (
	"slices"
	"time"
)

type Mock struct{}

func (Mock) Contains(id int) bool {
	visitedStories := []int{2, 10, 14, 18}

	return slices.Contains(visitedStories, id)
}

func (Mock) CommentsLastVisited(_ int) int64 {
	return time.Now().Unix()
}

func (Mock) Clear() error { return nil }

func (Mock) MarkRead(_ int, _ int) error { return nil }

func (Mock) MarkArticleRead(_ int) error { return nil }

func (Mock) MarkUnread(_ int) error { return nil }
