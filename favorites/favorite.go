package favorites

import (
	"clx/core"
	"clx/file"
	"encoding/json"
	"io/ioutil"
)

type Favorites struct {
	Items []*core.Submission
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

func unmarshal(data []byte) []*core.Submission {
	var subs []*core.Submission

	err := json.Unmarshal(data, &subs)
	if err != nil {
		panic(err)
	}

	return subs
}
