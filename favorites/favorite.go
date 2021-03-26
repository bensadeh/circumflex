package favorites

import (
	"clx/constants/settings"
	"clx/file"
	"encoding/json"
	"io/ioutil"
)

type Favorites struct {
	Items []Item
}

type Item struct {
	ID   int `json:"id"`
	Time int `json:"time"`
}

func Initialize() *Favorites {
	if !file.Exists(settings.FavoritesFilePath) {
		favs := new(Favorites)
		favs.Items = append(favs.Items, Item{1, 123456789})
		favs.Items = append(favs.Items, Item{2, 123456789})

		bytes, _ := json.Marshal(favs)

		_ = file.WriteToFile(settings.FavoritesFilePath, string(bytes))

		return favs
	}

	favoritesJSON, _ := ioutil.ReadFile(settings.FavoritesFilePath)

	favorites := unmarshal(favoritesJSON)

	return favorites
}

func unmarshal(data []byte) *Favorites {
	favorites := new(Favorites)

	err := json.Unmarshal(data, &favorites)
	if err != nil {
		panic(err)
	}

	return favorites
}
