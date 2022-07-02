package favorites

import (
	"encoding/json"
	"io/ioutil"

	"clx/file"
	"clx/item"
)

type Favorites struct {
	Items []*item.Item
}

func Initialize() *Favorites {
	favoritesPath := file.PathToFavoritesFile()

	if file.Exists(favoritesPath) {
		favoritesJSON, _ := ioutil.ReadFile(favoritesPath)
		items := unmarshal(favoritesJSON)

		favoritesFromDisk := new(Favorites)
		favoritesFromDisk.Items = items

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
