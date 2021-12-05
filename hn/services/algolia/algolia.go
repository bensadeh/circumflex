package algolia

import (
	"clx/constants/categories"
	"clx/constants/clx"
	"clx/item"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	uri                = "https://hacker-news.firebaseio.com/v0/"
	numberOfCategories = 3
)

type Service struct {
	numberOfItemsToShow int
	categories          []category
}

type category struct {
	items []int
}

func (s *Service) Init(itemsToShow int) {
	s.numberOfItemsToShow = itemsToShow
	s.categories = make([]category, numberOfCategories)
}

func (s *Service) FetchStories(page int, category int) []*item.Item {
	initializeStoriesList(s, category)

	// url := fmt.Sprintf("https://hn.algolia.com/api/v1/search?"+
	//	"tags=front_page"+
	//	"&hitsPerPage=%d"+
	//	"&page=%d", s.numberOfItemsToShow, page-1)
	ids := getStoryListURIParam(s.categories[category].items, s.numberOfItemsToShow, page)

	toShow := strconv.Itoa(s.numberOfItemsToShow)
	url := "https://hn.algolia.com/api/v1/search?tags=story," +
		"(" + ids + ")&hitsPerPage=" + toShow

	var a *algolia

	client := resty.New()
	client.SetTimeout(5 * time.Second)

	_, _ = client.R().
		SetHeader("User-Agent", clx.Name+"/"+clx.Version).
		SetResult(&a).
		Get(url)

	return mapStories(a)
}

func initializeStoriesList(s *Service, category int) {
	if len(s.categories[category].items) != 0 {
		return
	}

	s.categories[category].items = fetchStoriesList(category)
}

func fetchStoriesList(category int) []int {
	var stories []int

	url := fmt.Sprintf("%s/%s.json", uri, getCategory(category))

	client := resty.New()
	_, _ = client.R().
		SetHeader("User-Agent", clx.Name+"/"+clx.Version).
		SetResult(&stories).
		Get(url)
	//if err != nil {
	//	return stories, fmt.Errorf("could not fetch stories: %w", err)
	//}

	return stories
}

func getStoryListURIParam(ids []int, numberOfItemsToShow int, page int) string {
	var sb strings.Builder

	start := numberOfItemsToShow * page
	end := start + numberOfItemsToShow + 1

	for i := start; i < end; i++ {
		sb.WriteString(fmt.Sprintf("story_%d,", ids[i]))
	}

	return sb.String()
}

func getCategory(category int) string {
	switch category {
	case categories.FrontPage:
		return "topstories"

	case categories.New:
		return "newstories"

	case categories.Ask:
		return "askstories"

	case categories.Show:
		return "showstories"

	default:
		panic("Unsupported category: " + strconv.Itoa(category))
	}
}

func mapStories(stories *algolia) []*item.Item {
	// items := make([]*item.Item, 0, len(stories.Hits))

	var items []*item.Item

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
