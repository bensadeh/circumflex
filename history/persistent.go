package history

import (
	"encoding/json"
	"os"
	"time"
)

type Persistent struct {
	VisitedStories map[int]StoryInfo
}

type StoryInfo struct {
	LastVisited         int64
	CommentsOnLastVisit int
}

func (his *Persistent) Contains(id int) bool {
	_, contains := his.VisitedStories[id]

	return contains
}

func (his *Persistent) GetLastVisited(id int) int64 {
	if item, contains := his.VisitedStories[id]; contains {
		return item.LastVisited
	}

	return time.Now().Unix()
}

func (his *Persistent) GetLastCommentCount(id int) int {
	if item, contains := his.VisitedStories[id]; contains {
		return item.CommentsOnLastVisit
	}

	return 0
}

func (his *Persistent) ClearAndWriteToDisk() error {
	his.VisitedStories = make(map[int]StoryInfo)

	_, dirPath, fileName := getCacheFilePaths()
	return writeToDisk(his, dirPath, fileName)
}

func (his *Persistent) MarkAsReadAndWriteToDisk(id int, commentsOnLastVisit int) error {
	his.VisitedStories[id] = StoryInfo{
		LastVisited:         time.Now().Unix(),
		CommentsOnLastVisit: commentsOnLastVisit,
	}

	_, dirPath, fileName := getCacheFilePaths()
	return writeToDisk(his, dirPath, fileName)
}

func Initialize(isEnabled bool) *Persistent {
	h := &Persistent{
		VisitedStories: make(map[int]StoryInfo),
	}

	fullPath, dirPath, fileName := getCacheFilePaths()

	if !exists(fullPath) {
		_ = writeToDisk(h, dirPath, fileName)

		return h
	}

	historyFileContent, readErr := os.ReadFile(fullPath)
	if readErr != nil {
		return h
	}

	deserializationErr := json.Unmarshal(historyFileContent, &h.VisitedStories)
	if deserializationErr != nil {
		_ = h.ClearAndWriteToDisk()
	}

	return h
}
