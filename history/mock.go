package history

import (
	"slices"
	"time"
)

type Mock struct{}

func (his *Mock) NewHistory() *History {
	return nil
}

func (Mock) Contains(id int) bool {
	visitedStories := []int{2, 10, 14, 18}

	return slices.Contains(visitedStories, id)
}

func (Mock) GetLastVisited(_ int) int64 {
	return time.Now().Unix()
}

func (Mock) GetLastCommentCount(_ int) int {
	return 0
}

func (Mock) ClearAndWriteToDisk() error { return nil }

func (Mock) MarkAsReadAndWriteToDisk(_ int, _ int) error { return nil }
