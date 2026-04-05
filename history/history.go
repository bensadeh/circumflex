package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bensadeh/circumflex/fileutil"
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

	if !fileutil.Exists(filePath) {
		if err := writeToDisk(h, filePath); err != nil {
			return h, fmt.Errorf("could not create history at %s: %w", filePath, err)
		}

		return h, nil
	}

	historyFileContent, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return h, fmt.Errorf("could not read history at %s: %w", filePath, readErr)
	}

	deserializationErr := json.Unmarshal(historyFileContent, &h.VisitedStories)
	if deserializationErr != nil {
		return h, fmt.Errorf("history at %s is corrupted: %w\n  To start fresh, delete the file and restart", filePath, deserializationErr)
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

	return fileutil.WriteAtomic(filePath, string(visitedStoriesJSON))
}
