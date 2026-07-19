package favorites

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/fileutil"
	"github.com/bensadeh/circumflex/hn"
)

// sanitizeItem strips terminal escapes from an item's text fields. Items
// added in-session copy already-sanitized story fields, but a favorites file
// loaded or migrated from disk is untrusted input — a pre-sanitization 4.x
// favorites.json, or a hand-edited favorites.toml, can carry raw escapes.
func sanitizeItem(item *Item) {
	item.Title = ansi.Strip(item.Title)
	item.Author = ansi.Strip(item.Author)
	item.URL = ansi.Strip(item.URL)
	item.Domain = ansi.Strip(item.Domain)
}

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
type Item struct {
	ID            int    `toml:"id"`
	Title         string `toml:"title"`
	Points        int    `toml:"points"`
	Author        string `toml:"author"`
	Time          int64  `toml:"time"`
	URL           string `toml:"url"`
	Domain        string `toml:"domain"`
	CommentsCount int    `toml:"comments_count"`
}

// document is the top-level shape of favorites.toml: one [[favorites]]
// table per story.
type document struct {
	Favorites []*Item `toml:"favorites"`
}

type Favorites struct {
	mu    sync.RWMutex
	items []*Item
	path  string
}

// New loads favorites from path. When path does not exist but a pre-5.0
// favorites.json does, its items are converted and written to path once;
// the JSON file is left behind untouched.
func New(path, legacyJSONPath string) (*Favorites, error) {
	f := &Favorites{path: path}

	if fileutil.Exists(path) {
		items, err := load(path)
		if err != nil {
			return f, err
		}

		f.items = items

		return f, nil
	}

	if fileutil.Exists(legacyJSONPath) {
		items, err := loadLegacyJSON(legacyJSONPath)
		if err != nil {
			return f, err
		}

		f.items = items

		if err := f.Write(); err != nil {
			return f, fmt.Errorf("could not migrate favorites to %s: %w", path, err)
		}
	}

	return f, nil
}

func load(path string) ([]*Item, error) {
	var doc document

	md, err := toml.DecodeFile(path, &doc)
	if err != nil {
		return nil, fmt.Errorf("could not parse favorites (file may be corrupted): %w", err)
	}

	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, key := range undecoded {
			keys[i] = key.String()
		}

		return nil, fmt.Errorf("unknown keys in %s: %s", path, strings.Join(keys, ", "))
	}

	for _, item := range doc.Favorites {
		sanitizeItem(item)
	}

	return doc.Favorites, nil
}

// Items returns the internal slice. Callers on the same goroutine (Bubble Tea
// Update loop) can safely iterate it; concurrent use would need a copy.
func (f *Favorites) Items() []*Item {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.items
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
	var buf bytes.Buffer

	if err := toml.NewEncoder(&buf).Encode(document{Favorites: f.items}); err != nil {
		return fmt.Errorf("could not serialize favorites: %w", err)
	}

	if err := fileutil.WriteAtomic(f.path, buf.String()); err != nil {
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
