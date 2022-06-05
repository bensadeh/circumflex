package bfavorites

import (
	"clx/file"
	"clx/item"
	"encoding/json"
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

func (f Favorites) IsEmpty() bool {
	return len(f.items) == 0
}
