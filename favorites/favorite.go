package favorites

import (
	"clx/file"
	"clx/item"
	"encoding/json"
	"io/ioutil"
)

type Favorites struct {
	Items []*item.Item
}

func Initialize() *Favorites {
	favoritesPath := file.PathToFavoritesFile()

	if file.Exists(favoritesPath) {
		favoritesJSON, _ := ioutil.ReadFile(favoritesPath)
		subs := unmarshal(favoritesJSON)

		f := new(Favorites)
		f.Items = subs

		return f
	}

	f := new(Favorites)

	return f
}

func unmarshal(data []byte) []*item.Item {
	var subs []*item.Item

	err := json.Unmarshal(data, &subs)
	if err != nil {
		panic(err)
	}

	return subs
}
