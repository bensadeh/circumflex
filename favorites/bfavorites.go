package favorites

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bensadeh/circumflex/file"
	"github.com/bensadeh/circumflex/item"
)

type Favorites struct {
	items []*item.Item
}

func New() *Favorites {
	favoritesPath := file.PathToFavoritesFile()

	if file.Exists(favoritesPath) {
		favoritesJSON, _ := os.ReadFile(favoritesPath)
		items := unmarshal(favoritesJSON)

		favoritesFromDisk := new(Favorites)
		favoritesFromDisk.items = items

		return favoritesFromDisk
	}

	return new(Favorites)
}

func unmarshal(data []byte) []*item.Item {
	var items []*item.Item

	err := json.Unmarshal(data, &items)
	if err != nil {
		panic(err)
	}

	return items
}

func (f *Favorites) GetItems() []*item.Item {
	return f.items
}

func (f *Favorites) HasItems() bool {
	return len(f.items) != 0
}

func (f *Favorites) Add(item *item.Item) {
	f.items = append(f.items, item)
}

func (f *Favorites) Write() {
	err := file.WriteToFile(file.PathToFavoritesFile(), serializeToJson(f.items))
	if err != nil {
		panic(fmt.Errorf("could not write to file: %w", err))
	}
}

func serializeToJson(favorites []*item.Item) string {
	stream, err := json.MarshalIndent(favorites, "", "    ")
	if err != nil {
		panic(fmt.Errorf("could not serialize favorites struct: %w", err))
	}

	return string(stream)
}

func (f *Favorites) Remove(index int) {
	if index < 0 || index > len(f.items) {
		errorString := fmt.Sprintf("Out of bounds access for slice. Tried to remove index of %d, but size of "+
			"slice was %d", index, len(f.items))
		panic(errorString)
	}

	f.items = append(f.items[:index], f.items[index+1:]...)
}

func (f *Favorites) UpdateStoryAndWriteToDisk(newItem *item.Item) {
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

				f.Write()
			}
		}
	}
}
