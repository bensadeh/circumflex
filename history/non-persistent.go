package history

type NonPersistent struct{}

func (NonPersistent) Contains(id int) bool {
	//TODO implement me
	panic("implement me")
}

func (NonPersistent) GetLastVisited(id int) int64 {
	//TODO implement me
	panic("implement me")
}

func (NonPersistent) GetLastCommentCount(id int) int {
	//TODO implement me
	panic("implement me")
}

func (NonPersistent) ClearAndWriteToDisk() {
	//TODO implement me
	panic("implement me")
}

func (NonPersistent) AddToHistoryAndWriteToDisk(id int, commentsOnLastVisit int) {
	//TODO implement me
	panic("implement me")
}
