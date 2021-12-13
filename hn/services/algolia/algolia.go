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
	s.numberOfItemsToShow = (itemsToShow * 3) + 1
	s.categories = make([]category, numberOfCategories)
}

func (s *Service) FetchStories(_ int, category int) []*item.Item {
	initializeStoriesList(s, category)

	// url := fmt.Sprintf("https://hn.algolia.com/api/v1/search?"+
	//	"tags=front_page"+
	//	"&hitsPerPage=%d"+
	//	"&page=%d", s.numberOfItemsToShow, page-1)
	ids := getStoryListURIParam(s.categories[category].items)

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

	stories := fetchStoriesList(category)
	storiesSubset := shortenStories(stories, s.numberOfItemsToShow)

	s.categories[category].items = storiesSubset
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

func shortenStories(stories []int, storiesToShow int) []int {
	return stories[0:storiesToShow]
}

func getStoryListURIParam(ids []int) string {
	var sb strings.Builder

	for _, id := range ids {
		sb.WriteString(fmt.Sprintf("story_%d,", id))
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
	comments := new(comment)

	client := resty.New()
	_, _ = client.R().
		SetHeader("User-Agent", clx.Name+"/"+clx.Version).
		SetResult(&comments).
		Get("https://hn.algolia.com/api/v1/items/" + strconv.Itoa(id))

	//if err != nil {
	//	return stories, fmt.Errorf("could not fetch stories: %w", err)
	//}

	return mapComments(comments, -1)
}

func mapComments(comments *comment, level int) *item.Item {
	items := make([]*item.Item, 0, len(comments.Children))

	for i := range comments.Children {
		items = append(items, mapComments(comments.Children[i], level+1))
	}

	return &item.Item{
		ID:            comments.ID,
		Title:         comments.Title,
		Points:        comments.Points,
		User:          comments.Author,
		Time:          int64(comments.CreatedAtI),
		TimeAgo:       "",
		Type:          comments.Type,
		URL:           comments.URL,
		Level:         level,
		Domain:        "",
		Comments:      items,
		Content:       comments.Text,
		CommentsCount: 0,
	}
}
