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
	Clear() error
	MarkRead(id int, commentsOnLastVisit int) error
	MarkArticleRead(id int) error
	MarkUnread(id int) error
}

func NewPersistentHistory() (History, error) {
	filePath := filepath.Join(settings.CachePath(), "history.json")

	h := &Persistent{
		filePath:       filePath,
		visitedStories: make(map[int]StoryInfo),
	}

	if !fileutil.Exists(filePath) {
		if err := h.writeToDisk(); err != nil {
			return h, fmt.Errorf("could not create history at %s: %w", filePath, err)
		}

		return h, nil
	}

	historyFileContent, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return h, fmt.Errorf("could not read history at %s: %w", filePath, readErr)
	}

	deserializationErr := json.Unmarshal(historyFileContent, &h.visitedStories)
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

func (his *Persistent) writeToDisk() error {
	visitedStoriesJSON, err := json.Marshal(his.visitedStories)
	if err != nil {
		return err
	}

	return fileutil.WriteAtomic(his.filePath, string(visitedStoriesJSON))
}
