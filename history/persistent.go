package history

import (
	"sync"
	"time"
)

type Persistent struct {
	mu             sync.RWMutex `json:"-"`
	filePath       string       `json:"-"`
	VisitedStories map[int]StoryInfo
}

type StoryInfo struct {
	LastVisited         int64
	CommentsOnLastVisit int
}

func (his *Persistent) Contains(id int) bool {
	his.mu.RLock()
	defer his.mu.RUnlock()

	_, contains := his.VisitedStories[id]

	return contains
}

func (his *Persistent) GetLastVisited(id int) int64 {
	his.mu.RLock()
	defer his.mu.RUnlock()

	if item, contains := his.VisitedStories[id]; contains {
		return item.LastVisited
	}

	return time.Now().Unix()
}

func (his *Persistent) ClearAndWriteToDisk() error {
	his.mu.Lock()
	defer his.mu.Unlock()

	his.VisitedStories = make(map[int]StoryInfo)

	return writeToDisk(his, his.filePath)
}

func (his *Persistent) MarkAsReadAndWriteToDisk(id int, commentsOnLastVisit int) error {
	his.mu.Lock()
	defer his.mu.Unlock()

	if existing, ok := his.VisitedStories[id]; ok {
		elapsed := time.Since(time.Unix(existing.LastVisited, 0))
		if elapsed < 5*time.Minute {
			return nil
		}
	}

	his.VisitedStories[id] = StoryInfo{
		LastVisited:         time.Now().Unix(),
		CommentsOnLastVisit: commentsOnLastVisit,
	}

	return writeToDisk(his, his.filePath)
}

func (his *Persistent) MarkAsUnreadAndWriteToDisk(id int) error {
	his.mu.Lock()
	defer his.mu.Unlock()

	delete(his.VisitedStories, id)

	return writeToDisk(his, his.filePath)
}
