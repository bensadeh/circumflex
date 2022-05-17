package history

type Mock struct{}

func (Mock) Contains(id int) bool {
	visitedStories := []int{1, 2, 3}

	return contains(visitedStories, id)
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (Mock) GetLastVisited(id int) int64 {
	//TODO implement me
	panic("implement me")
}

func (Mock) GetLastCommentCount(id int) int {
	//TODO implement me
	panic("implement me")
}

func (Mock) ClearAndWriteToDisk() {
	//TODO implement me
	panic("implement me")
}

func (Mock) AddToHistoryAndWriteToDisk(id int, commentsOnLastVisit int) {
	//TODO implement me
	panic("implement me")
}
