package history

import (
	"clx/file"
	"encoding/json"
	"os"
	"path"
)

type History interface {
	Contains(id int) bool
	CommentsLastVisited(id int) int64
	ClearAndWriteToDisk() error
	MarkAsReadAndWriteToDisk(id int, commentsOnLastVisit int) error
	MarkArticleAsReadAndWriteToDisk(id int) error
	MarkAsUnreadAndWriteToDisk(id int) error
}

func NewPersistentHistory() (History, error) {
	filePath := path.Join(file.PathToCacheDirectory(), "history.json")

	h := &Persistent{
		filePath:       filePath,
		VisitedStories: make(map[int]StoryInfo),
	}

	if !file.Exists(filePath) {
		if err := writeToDisk(h, filePath); err != nil {
			return h, err
		}

		return h, nil
	}

	historyFileContent, readErr := os.ReadFile(filePath)
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

func writeToDisk(h *Persistent, filePath string) error {
	visitedStoriesJSON, err := json.Marshal(h.VisitedStories)
	if err != nil {
		return err
	}

	return file.WriteToDir(path.Dir(filePath), path.Base(filePath), string(visitedStoriesJSON))
}
