package history

import (
	"clx/file"
	"os"
	"path"
	"strconv"

	"github.com/emirpasic/gods/sets/hashset"
)

type History struct {
	visitedStories *hashset.Set
	markAsRead     bool
}

func (his *History) Contains(id int) bool {
	if !his.markAsRead {
		return false
	}

	return his.visitedStories.Contains(strconv.Itoa(id))
}

func (his *History) AddStoryAndWriteToDisk(id int) {
	if !his.markAsRead {
		return
	}

	his.visitedStories.Add(strconv.Itoa(id))

	_, dirPath, fileName := getCacheFilePaths()
	writeToDisk(his, dirPath, fileName)
}

func Initialize(markAsRead bool) *History {
	h := &History{
		visitedStories: hashset.New(),
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

	deserializationErr := h.visitedStories.FromJSON(historyFileContent)
	if deserializationErr != nil {
		panic(deserializationErr)
	}

	return h
}

func writeToDisk(h *History, dirPath string, fileName string) {
	visitedStoriesJSON, _ := h.visitedStories.ToJSON()

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
