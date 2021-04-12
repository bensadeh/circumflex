package favorites

import (
	"clx/endpoints"
	"clx/file"
	"encoding/json"
	"io/ioutil"
)

type Favorites struct {
	Items []*endpoints.Story
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

func unmarshal(data []byte) []*endpoints.Story {
	var subs []*endpoints.Story

	err := json.Unmarshal(data, &subs)
	if err != nil {
		panic(err)
	}

	return subs
}
