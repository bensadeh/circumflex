package favorites

import (
	"clx/item"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Favorites struct {
	items []*item.Story
	path  string
}

func New(path string) *Favorites {
	f := &Favorites{path: path}

	if fileExists(path) {
		favoritesJSON, err := os.ReadFile(path)
		if err != nil {
			return f
		}

		items, err := unmarshal(favoritesJSON)
		if err != nil {
			return f
		}

		f.items = items
	}

	return f
}

func unmarshal(data []byte) ([]*item.Story, error) {
	var items []*item.Story

	err := json.Unmarshal(data, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (f *Favorites) Items() []*item.Story {
	return f.items
}

func (f *Favorites) HasItems() bool {
	return len(f.items) != 0
}

func (f *Favorites) Add(item *item.Story) {
	f.items = append(f.items, item)
}

func (f *Favorites) Write() error {
	stream, err := json.MarshalIndent(f.items, "", "    ")
	if err != nil {
		return fmt.Errorf("could not serialize favorites: %w", err)
	}

	if err := writeFile(f.path, string(stream)); err != nil {
		return fmt.Errorf("could not write favorites: %w", err)
	}

	return nil
}

func (f *Favorites) Remove(index int) error {
	if index < 0 || index >= len(f.items) {
		return fmt.Errorf("out of bounds: tried to remove index %d, but size was %d", index, len(f.items))
	}

	f.items = append(f.items[:index], f.items[index+1:]...)

	return nil
}

func (f *Favorites) UpdateStoryAndWriteToDisk(newItem *item.Story) error {
	for i, s := range f.items {
		if s.ID == newItem.ID {
			isFieldsUpdated := s.Title != newItem.Title || s.Points != newItem.Points ||
				s.Time != newItem.Time || s.User != newItem.User ||
				s.CommentsCount != newItem.CommentsCount || s.URL != newItem.URL ||
				s.Domain != newItem.Domain

			if isFieldsUpdated {
				f.items[i].Title = newItem.Title
				f.items[i].Points = newItem.Points
				f.items[i].Time = newItem.Time
				f.items[i].User = newItem.User
				f.items[i].CommentsCount = newItem.CommentsCount
				f.items[i].URL = newItem.URL
				f.items[i].Domain = newItem.Domain

				return f.Write()
			}
		}
	}

	return nil
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
