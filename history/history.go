package history

import (
	"clx/file"
	"encoding/json"
	"os"
	"path"

	"github.com/emirpasic/gods/sets/hashset"
)

const (
	disableHistory = 0
)

type Handler struct {
	visitedStories *hashset.Set
	mode           int
}

func (h *Handler) Initialize(historyMode int) {
	h.mode = historyMode

	if h.mode == disableHistory {
		h.visitedStories = hashset.New()

		return
	}

	fullPath, dirPath, fileName := getCacheFilePaths()

	if !exists(fullPath) {
		writeToDisk(h, dirPath, fileName)

		return
	}

	historyFileContent, err := os.ReadFile(fullPath)
	if err != nil {
		panic(err)
	}

	h.visitedStories = unmarshal(historyFileContent)
}

func writeToDisk(h *Handler, dirPath string, fileName string) {
	h.visitedStories = hashset.New()
	emptyJSON, _ := h.visitedStories.ToJSON()

	err := file.WriteToFileNew(dirPath, fileName, string(emptyJSON))
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

func unmarshal(input []byte) *hashset.Set {
	cache := hashset.New()

	err := json.Unmarshal(input, &cache)
	if err != nil {
		panic(err)
	}

	return cache
}
