package bfavorites

import (
	"clx/file"
	"clx/item"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Favorites struct {
	items []*item.Item
}

func New() *Favorites {
	favoritesPath := file.PathToFavoritesFile()

	if file.Exists(favoritesPath) {
		favoritesJSON, _ := ioutil.ReadFile(favoritesPath)
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

func (f Favorites) GetItems() []*item.Item {
	return f.items
}

func (f Favorites) HasItems() bool {
	return len(f.items) != 0
}

func (f Favorites) Add(item *item.Item) {
	f.items = append(f.items, item)
}

func (f Favorites) Write() {
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
