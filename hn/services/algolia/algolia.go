package algolia

import (
	"clx/constants/clx"
	"clx/item"
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

type Service struct {
	numberOfItemsToShow int
}

func (s *Service) Init(itemsToShow int) {
	s.numberOfItemsToShow = itemsToShow
}

const (
	baseURL = "https://hn.algolia.com/api/v1/search?tags=front_page"
	page    = "?page="
)

func (s *Service) FetchStories(page int, category int) []*item.Item {
	// url := getURL(category)
	// p := strconv.Itoa(page)
	url := fmt.Sprintf("https://hn.algolia.com/api/v1/search?"+
		"tags=front_page"+
		"&hitsPerPage=%d"+
		"&page=%d", s.numberOfItemsToShow, page-1)
	var a *algolia

	client := resty.New()
	client.SetTimeout(5 * time.Second)

	_, _ = client.R().
		SetHeader("User-Agent", clx.Name+"/"+clx.Version).
		SetResult(&a).
		Get(url)

	return mapStories(a)
}

func mapStories(stories *algolia) []*item.Item {
	items := make([]*item.Item, 0, len(stories.Hits))

	for _, story := range stories.Hits {
		id, _ := strconv.Atoi(story.ObjectID)

		item := item.Item{
			ID:            id,
			Title:         story.Title,
			Points:        story.Points,
			User:          story.Author,
			Time:          int64(story.CreatedAtI),
			TimeAgo:       "",
			Type:          "",
			URL:           story.URL,
			Domain:        "",
			Comments:      nil,
			Content:       "",
			Level:         0,
			CommentsCount: story.NumComments,
		}

		items = append(items, &item)
	}

	return items
}

func (s *Service) FetchStory(id int) *item.Item {
	return nil
}
