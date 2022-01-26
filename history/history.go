package history

import (
	"clx/file"
	"encoding/json"
	"os"
	"path"
	"time"
)

type History struct {
	VisitedStories map[int]Data
	markAsRead     bool
}

type Data struct {
	LastVisited         int64
	CommentsOnLastVisit int
}

func (his *History) Contains(id int) bool {
	if !his.markAsRead {
		return false
	}

	_, contains := his.VisitedStories[id]

	return contains
}

func (his *History) GetLastVisited(id int) int64 {
	if !his.markAsRead {
		return time.Now().Unix()
	}

	if item, contains := his.VisitedStories[id]; contains {
		return item.LastVisited
	}

	return time.Now().Unix()
}

func (his *History) GetLastCommentCount(id int) int {
	if !his.markAsRead {
		return 0
	}

	if item, contains := his.VisitedStories[id]; contains {
		return item.CommentsOnLastVisit
	}

	return 0
}

func (his *History) ClearAndWriteToDisk() {
	his.VisitedStories = make(map[int]Data)

	_, dirPath, fileName := getCacheFilePaths()
	writeToDisk(his, dirPath, fileName)
}

func (his *History) AddToHistoryAndWriteToDisk(id int, commentsOnLastVisit int) {
	if !his.markAsRead {
		return
	}

	his.VisitedStories[id] = Data{
		LastVisited:         time.Now().Unix(),
		CommentsOnLastVisit: commentsOnLastVisit,
	}

	_, dirPath, fileName := getCacheFilePaths()
	writeToDisk(his, dirPath, fileName)
}

func Initialize(markAsRead bool) *History {
	h := &History{
		VisitedStories: make(map[int]Data),
		markAsRead:     markAsRead,
	}

	if !h.markAsRead {
		return h
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

	deserializationErr := json.Unmarshal(historyFileContent, &h.VisitedStories)
	if deserializationErr != nil {
		h.ClearAndWriteToDisk()
		_ = json.Unmarshal(historyFileContent, &h.VisitedStories)
	}

	return h
}

func writeToDisk(h *History, dirPath string, fileName string) {
	visitedStoriesJSON, _ := json.Marshal(h.VisitedStories)

	err := file.WriteToFileNew(dirPath, fileName, string(visitedStoriesJSON))
	if err != nil {
		panic(err)
	}
}

func getCacheFilePaths() (string, string, string) {
	homeDir, _ := os.UserHomeDir()
	configDir := ".cache"
	circumflexDir := "circumflex"
	fileName := "history.json"

	fullPath := path.Join(homeDir, configDir, circumflexDir, fileName)
	dirPath := path.Join(homeDir, configDir, circumflexDir)

	return fullPath, dirPath, fileName
}

func exists(pathToFile string) bool {
	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return false
	}

	return true
}
