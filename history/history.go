package history

import (
	"clx/file"
	"encoding/json"
	"os"
	"path"
)

type History interface {
	Contains(id int) bool
	GetLastVisited(id int) int64
	ClearAndWriteToDisk() error
	MarkAsReadAndWriteToDisk(id int, commentsOnLastVisit int) error
}

func NewPersistentHistory() (History, error) {
	h := &Persistent{VisitedStories: make(map[int]StoryInfo)}

	fullPath, dirPath, fileName := getCacheFilePaths()

	if !file.Exists(fullPath) {
		if err := writeToDisk(h, dirPath, fileName); err != nil {
			return h, err
		}

		return h, nil
	}

	historyFileContent, readErr := os.ReadFile(fullPath) //nolint:gosec // path from ~/.cache/circumflex/
	if readErr != nil {
		// Graceful degradation: treat as empty history
		return h, nil //nolint:nilerr
	}

	deserializationErr := json.Unmarshal(historyFileContent, &h.VisitedStories)
	if deserializationErr != nil {
		if clearErr := h.ClearAndWriteToDisk(); clearErr != nil {
			return h, clearErr
		}
	}

	return h, nil
}

func NewNonPersistentHistory() History {
	return &NonPersistent{}
}

func NewMockHistory() History {
	return &Mock{}
}

func writeToDisk(h *Persistent, dirPath string, fileName string) error {
	visitedStoriesJSON, err := json.Marshal(h.VisitedStories)
	if err != nil {
		return err
	}

	return file.WriteToDir(dirPath, fileName, string(visitedStoriesJSON))
}

func getCacheFilePaths() (string, string, string) {
	dirPath := file.PathToCacheDirectory()
	fileName := "history.json"
	fullPath := path.Join(dirPath, fileName)

	return fullPath, dirPath, fileName
}
