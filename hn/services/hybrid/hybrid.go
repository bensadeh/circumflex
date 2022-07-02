package hybrid

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"clx/app"
	"clx/constants/category"
	"clx/endpoints"
	"clx/item"

	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/go-resty/resty/v2"
)

const (
	uri = "https://hacker-news.firebaseio.com/v0"
)

type Service struct{}

func (s *Service) FetchItems(itemsToFetch int, category int) []*item.Item {
	// Posts of the type: 'Company (YC __) is hiring ...' is filtered out
	// from Algolia. For this reason, we ask for one more item than we need.
	itemsToFetchWithBuffer := itemsToFetch + 1
	listOfIDs := fetchStoriesList(category)

	ids := getStoryListURIParam(listOfIDs[0:itemsToFetchWithBuffer])

	url := "https://hn.algolia.com/api/v1/search?tags=story," +
		"(" + ids + ")&hitsPerPage=" + strconv.Itoa(itemsToFetchWithBuffer)

	var a *endpoints.Algolia

	client := resty.New()
	client.SetTimeout(5 * time.Second)

	_, _ = client.R().
		SetHeader("User-Agent", app.Name+"/"+app.Version).
		SetResult(&a).
		Get(url)

	mapOfItemsWithMetaData := mapStories(a)
	orderedStories := joinStories(listOfIDs, mapOfItemsWithMetaData)

	return orderedStories[0:min(itemsToFetch, len(orderedStories))]
}

func fetchStoriesList(category int) []int {
	var stories []int

	url := fmt.Sprintf("%s/%s.json", uri, getCategory(category))

	client := resty.New()
	client.SetTimeout(5 * time.Second)

	_, _ = client.R().
		SetHeader("User-Agent", app.Name+"/"+app.Version).
		SetResult(&stories).
		Get(url)

	return stories
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

func mapStories(stories *endpoints.Algolia) map[int]*item.Item {
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

func (s Service) FetchItem(id int) *item.Item {
	hn := new(endpoints.HN)

	client := resty.New()
	client.SetTimeout(5 * time.Second)
	client.SetBaseURL("https://hacker-news.firebaseio.com/v0/item/")

	_, _ = client.R().
		SetHeader("User-Agent", app.Name+"/"+app.Version).
		SetResult(hn).
		Get(strconv.Itoa(id) + ".json")

	return mapItem(hn)
}

func mapItem(hn *endpoints.HN) *item.Item {
	return &item.Item{
		ID:            hn.Id,
		Title:         hn.Title,
		Points:        hn.Score,
		User:          hn.By,
		Time:          int64(hn.Time),
		TimeAgo:       "",
		Type:          "",
		URL:           hn.Url,
		CommentsCount: hn.Descendants,
	}
}

func (s Service) FetchComments(id int) *item.Item {
	comments := new(endpoints.Comments)

	client := resty.New()
	client.SetTimeout(5 * time.Second)
	client.SetBaseURL("http://api.hackerwebapp.com/item/")

	_, _ = client.R().
		SetHeader("User-Agent", app.Name+"/"+app.Version).
		SetResult(comments).
		Get(strconv.Itoa(id))

	return mapComments(comments)
}

func mapComments(comments *endpoints.Comments) *item.Item {
	items := make([]*item.Item, 0, len(comments.Comments))

	for i := range comments.Comments {
		items = append(items, mapComments(&comments.Comments[i]))
	}

	return &item.Item{
		ID:            comments.ID,
		Title:         comments.Title,
		Points:        comments.Points,
		User:          comments.User,
		Time:          comments.Time,
		TimeAgo:       comments.TimeAgo,
		Type:          comments.Type,
		URL:           comments.URL,
		Level:         comments.Level,
		Domain:        comments.Domain,
		Comments:      items,
		Content:       comments.Content,
		CommentsCount: comments.CommentsCount,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
