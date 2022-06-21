package algolia_bubble

import (
	"clx/constants/category"
	"clx/constants/clx"
	"clx/item"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bobesa/go-domain-util/domainutil"

	"github.com/go-resty/resty/v2"
)

const (
	uri = "https://hacker-news.firebaseio.com/v0"
)

type Service struct {
}

func (s *Service) Init(itemsToShow int) {

}

func (s *Service) FetchStories(itemsToFetch int, category int) []*item.Item {
	listOfIDs := fetchStoriesList(category)

	ids := getStoryListURIParam(listOfIDs[0:itemsToFetch])

	url := "https://hn.algolia.com/api/v1/search?tags=story," +
		"(" + ids + ")&hitsPerPage=" + strconv.Itoa(itemsToFetch)

	var a *algolia

	client := resty.New()
	client.SetTimeout(5 * time.Second)

	_, _ = client.R().
		SetHeader("User-Agent", clx.Name+"/"+clx.Version).
		SetResult(&a).
		Get(url)

	mapOfItemsWithMetaData := mapStories(a)
	orderedStories := joinStories(listOfIDs, mapOfItemsWithMetaData)

	return orderedStories
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
	return stories[0 : storiesToShow-1]
}

func getStoryListURIParam(ids []int) string {
	var sb strings.Builder

	for _, id := range ids {
		sb.WriteString(fmt.Sprintf("story_%d,", id))
	}

	return sb.String()
}

func getCategory(cat int) string {
	switch cat {
	case category.FrontPage:
		return "topstories"

	case category.New:
		return "newstories"

	case category.Ask:
		return "askstories"

	case category.Show:
		return "showstories"

	default:
		panic("Unsupported c: " + strconv.Itoa(cat))
	}
}

func mapStories(stories *algolia) map[int]*item.Item {
	m := make(map[int]*item.Item)

	for _, story := range stories.Hits {
		id, _ := strconv.Atoi(story.ObjectID)

		it := &item.Item{
			ID:            id,
			Title:         story.Title,
			Points:        story.Points,
			User:          story.Author,
			Time:          int64(story.CreatedAtI),
			TimeAgo:       "",
			Type:          "",
			URL:           story.URL,
			Domain:        domainutil.Domain(story.URL),
			Comments:      nil,
			Content:       "",
			Level:         0,
			CommentsCount: story.NumComments,
		}

		m[id] = it
	}

	return m
}

func joinStories(orderedIds []int, stories map[int]*item.Item) []*item.Item {
	var orderedStories []*item.Item

	for _, id := range orderedIds {
		if stories[id] == nil {
			continue
		}

		orderedStories = append(orderedStories, stories[id])
	}

	return orderedStories
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
