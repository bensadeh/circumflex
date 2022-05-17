package history

import (
	"encoding/json"
	"os"
	"time"
)

type Persistent struct {
	visitedStories map[int]storyInfo
}

type storyInfo struct {
	lastVisited         int64
	commentsOnLastVisit int
}

func (his *Persistent) Contains(id int) bool {
	_, contains := his.visitedStories[id]

	return contains
}

func (his *Persistent) GetLastVisited(id int) int64 {
	if item, contains := his.visitedStories[id]; contains {
		return item.lastVisited
	}

	return time.Now().Unix()
}

func (his *Persistent) GetLastCommentCount(id int) int {
	if item, contains := his.visitedStories[id]; contains {
		return item.commentsOnLastVisit
	}

	return 0
}

func (his *Persistent) ClearAndWriteToDisk() {
	his.visitedStories = make(map[int]storyInfo)

	_, dirPath, fileName := getCacheFilePaths()
	writeToDisk(his, dirPath, fileName)
}

func (his *Persistent) AddToHistoryAndWriteToDisk(id int, commentsOnLastVisit int) {
	his.visitedStories[id] = storyInfo{
		lastVisited:         time.Now().Unix(),
		commentsOnLastVisit: commentsOnLastVisit,
	}

	_, dirPath, fileName := getCacheFilePaths()
	writeToDisk(his, dirPath, fileName)
}

func Initialize(isEnabled bool) *Persistent {
	h := &Persistent{
		visitedStories: make(map[int]storyInfo),
	}

	fullPath, dirPath, fileName := getCacheFilePaths()

	if !exists(fullPath) {
		writeToDisk(h, dirPath, fileName)

		return h
	}

	historyFileContent, readErr := os.ReadFile(fullPath)
	if readErr != nil {
		panic(readErr)
	}

	deserializationErr := json.Unmarshal(historyFileContent, &h.visitedStories)
	if deserializationErr != nil {
		h.ClearAndWriteToDisk()
		_ = json.Unmarshal(historyFileContent, &h.visitedStories)
	}

	return h
}
