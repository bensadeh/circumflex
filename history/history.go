package history

import (
	"encoding/json"
	"os"
	"path"

	"github.com/f01c33/clx/file"
)

type History interface {
	Contains(id int) bool
	GetLastVisited(id int) int64
	GetLastCommentCount(id int) int
	ClearAndWriteToDisk()
	MarkAsReadAndWriteToDisk(id int, commentsOnLastVisit int)
}

func NewPersistentHistory() History {
	h := &Persistent{VisitedStories: make(map[int]StoryInfo)}

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

func NewNonPersistentHistory() History {
	return &NonPersistent{}
}

func NewMockHistory() History {
	return &Mock{}
}

func writeToDisk(h *Persistent, dirPath string, fileName string) {
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
