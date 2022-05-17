package history

import "time"

type Mock struct{}

func (his *Mock) NewHistory() *History {
	return nil
}

func (Mock) Contains(id int) bool {
	visitedStories := []int{1, 2, 3}

	return contains(visitedStories, id)
}

func contains(slice []int, element int) bool {
	for _, a := range slice {
		if a == element {
			return true
		}
	}

	return false
}

func (Mock) GetLastVisited(_ int) int64 {
	return time.Now().Unix()
}

func (Mock) GetLastCommentCount(_ int) int {
	return 0
}

func (Mock) ClearAndWriteToDisk() {}

func (Mock) AddToHistoryAndWriteToDisk(_ int, _ int) {}
