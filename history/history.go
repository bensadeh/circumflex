package history

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/bensadeh/circumflex/settings"
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
	filePath := filepath.Join(settings.CachePath(), "history.json")

	h := &Persistent{
		filePath:       filePath,
		VisitedStories: make(map[int]StoryInfo),
	}

	if !fileExists(filePath) {
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

	return writeFile(filePath, string(visitedStoriesJSON))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func writeFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0o600)
}
