package history

import (
	"sync"
	"time"
)

const readCooldown = 5 * time.Minute

type Persistent struct {
	mu             sync.RWMutex
	filePath       string
	visitedStories map[int]StoryInfo
}

type StoryInfo struct {
	LastVisited         int64
	CommentsLastVisited int64
	CommentsOnLastVisit int
}

func (his *Persistent) Contains(id int) bool {
	his.mu.RLock()
	defer his.mu.RUnlock()

	_, contains := his.visitedStories[id]

	return contains
}

func (his *Persistent) CommentsLastVisited(id int) int64 {
	his.mu.RLock()
	defer his.mu.RUnlock()

	if item, contains := his.visitedStories[id]; contains {
		if item.CommentsLastVisited > 0 {
			return item.CommentsLastVisited
		}

		return item.LastVisited
	}

	return time.Now().Unix()
}

func (his *Persistent) Clear() error {
	his.mu.Lock()
	defer his.mu.Unlock()

	his.visitedStories = make(map[int]StoryInfo)

	return his.writeToDisk()
}

func (his *Persistent) MarkRead(id int, commentsOnLastVisit int) error {
	his.mu.Lock()
	defer his.mu.Unlock()

	if existing, ok := his.visitedStories[id]; ok {
		elapsed := time.Since(time.Unix(existing.CommentsLastVisited, 0))
		if elapsed < readCooldown {
			return nil
		}
	}

	now := time.Now().Unix()

	his.visitedStories[id] = StoryInfo{
		LastVisited:         now,
		CommentsLastVisited: now,
		CommentsOnLastVisit: commentsOnLastVisit,
	}

	return his.writeToDisk()
}

func (his *Persistent) MarkArticleRead(id int) error {
	his.mu.Lock()
	defer his.mu.Unlock()

	now := time.Now().Unix()

	if existing, ok := his.visitedStories[id]; ok {
		elapsed := time.Since(time.Unix(existing.LastVisited, 0))
		if elapsed < readCooldown {
			return nil
		}

		his.visitedStories[id] = StoryInfo{
			LastVisited:         now,
			CommentsLastVisited: existing.CommentsLastVisited,
			CommentsOnLastVisit: existing.CommentsOnLastVisit,
		}
	} else {
		his.visitedStories[id] = StoryInfo{
			LastVisited: now,
		}
	}

	return his.writeToDisk()
}

func (his *Persistent) MarkUnread(id int) error {
	his.mu.Lock()
	defer his.mu.Unlock()

	delete(his.visitedStories, id)

	return his.writeToDisk()
}
