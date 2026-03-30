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

func (Mock) ClearAndWriteToDisk() error { return nil }

func (Mock) MarkAsReadAndWriteToDisk(_ int, _ int) error { return nil }

func (Mock) MarkArticleAsReadAndWriteToDisk(_ int) error { return nil }

func (Mock) MarkAsUnreadAndWriteToDisk(_ int) error { return nil }
