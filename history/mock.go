package history

import "slices"

import "time"

type Mock struct{}

func (his *Mock) NewHistory() *History {
	return nil
}

func (Mock) Contains(id int) bool {
	visitedStories := []int{2, 10, 14, 18}

	return contains(visitedStories, id)
}

func contains(slice []int, element int) bool {
	return slices.Contains(slice, element)
}

func (Mock) GetLastVisited(_ int) int64 {
	return time.Now().Unix()
}

func (Mock) GetLastCommentCount(_ int) int {
	return 0
}

func (Mock) ClearAndWriteToDisk() {}

func (Mock) MarkAsReadAndWriteToDisk(_ int, _ int) {}
