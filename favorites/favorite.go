package favorites

import (
	"clx/constants/settings"
	"clx/core"
	"clx/file"
	"encoding/json"
	"io/ioutil"
)

type Favorites struct {
	Items []*core.Submission
}

func Initialize() *Favorites {
	if !file.Exists(settings.FavoritesFilePath) {
		f := new(Favorites)
		f.Items = append(f.Items, &core.Submission{ID: 1, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "Title", Points: 2, Author: "author", Time: "1 second", CommentsCount: 2, URL: "google.com", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})
		f.Items = append(f.Items, &core.Submission{ID: 2, Title: "", Points: 2, Author: "", Time: "", CommentsCount: 2, URL: "", Domain: "", Type: ""})

		return f
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
