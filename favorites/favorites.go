package favorites

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/bensadeh/circumflex/fileutil"
	"github.com/bensadeh/circumflex/hn"
)

// ItemFromStory maps an hn.Story to a favorites Item.
func ItemFromStory(s *hn.Story) *Item {
	return &Item{
		ID:            s.ID,
		Title:         s.Title,
		Points:        s.Points,
		Author:        s.Author,
		Time:          s.Time,
		URL:           s.URL,
		Domain:        s.Domain,
		CommentsCount: s.CommentsCount,
	}
}

// Item holds the fields persisted for a favorited story.
// JSON tags match the old item.Story PascalCase names for backward compatibility.
type Item struct {
	ID            int    `json:"ID"`
	Title         string `json:"Title"`
	Points        int    `json:"Points"`
	Author        string `json:"User"`
	Time          int64  `json:"Time"`
	URL           string `json:"URL"`
	Domain        string `json:"Domain"`
	CommentsCount int    `json:"CommentsCount"`
}

type Favorites struct {
	mu    sync.RWMutex
	items []*Item
	path  string
}

func New(path string) (*Favorites, error) {
	f := &Favorites{path: path}

	if !fileutil.Exists(path) {
		return f, nil
	}

	favoritesJSON, err := os.ReadFile(path)
	if err != nil {
		return f, fmt.Errorf("could not read favorites: %w", err)
	}

	items, err := unmarshal(favoritesJSON)
	if err != nil {
		return f, fmt.Errorf("could not parse favorites (file may be corrupted): %w", err)
	}

	f.items = items

	return f, nil
}

func unmarshal(data []byte) ([]*Item, error) {
	var items []*Item

	err := json.Unmarshal(data, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// Items returns the internal slice. Callers on the same goroutine (Bubble Tea
// Update loop) can safely iterate it; concurrent use would need a copy.
func (f *Favorites) Items() []*Item {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.items
}

func (f *Favorites) HasItems() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.items) != 0
}

func (f *Favorites) Add(item *Item) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.items = append(f.items, item)
}

func (f *Favorites) Write() error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.writeLocked()
}

func (f *Favorites) writeLocked() error {
	stream, err := json.MarshalIndent(f.items, "", "    ")
	if err != nil {
		return fmt.Errorf("could not serialize favorites: %w", err)
	}

	if err := fileutil.WriteAtomic(f.path, string(stream)); err != nil {
		return fmt.Errorf("could not write favorites: %w", err)
	}

	return nil
}

func (f *Favorites) Remove(index int) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if index < 0 || index >= len(f.items) {
		return fmt.Errorf("out of bounds: tried to remove index %d, but size was %d", index, len(f.items))
	}

	f.items = append(f.items[:index], f.items[index+1:]...)

	return nil
}

func (f *Favorites) RemoveLast() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.items) == 0 {
		return fmt.Errorf("cannot remove from empty favorites")
	}

	f.items = f.items[:len(f.items)-1]

	return nil
}

func (f *Favorites) UpdateStoryAndWriteToDisk(newItem *Item) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i, s := range f.items {
		if s.ID == newItem.ID {
			isFieldsUpdated := s.Title != newItem.Title || s.Points != newItem.Points ||
				s.Time != newItem.Time || s.Author != newItem.Author ||
				s.CommentsCount != newItem.CommentsCount || s.URL != newItem.URL ||
				s.Domain != newItem.Domain

			if isFieldsUpdated {
				f.items[i].Title = newItem.Title
				f.items[i].Points = newItem.Points
				f.items[i].Time = newItem.Time
				f.items[i].Author = newItem.Author
				f.items[i].CommentsCount = newItem.CommentsCount
				f.items[i].URL = newItem.URL
				f.items[i].Domain = newItem.Domain

				return f.writeLocked()
			}
		}
	}

	return nil
}
