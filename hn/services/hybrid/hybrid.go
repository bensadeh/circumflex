package hybrid

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	ansi "clx/utils/strip-ansi"

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

func (s *Service) FetchItems(itemsToFetch int, category int) (items []*item.Item, errMsg string) {
	// Posts of the type: 'Company (YC __) is hiring ...' is filtered out
	// from Algolia. For this reason, we ask for one more item than we need.
	itemsToFetchWithBuffer := itemsToFetch + 1
	listOfIDs, errMsg := fetchStoriesList(category)
	if errMsg != "" {
		return nil, errMsg
	}

	ids := getStoryListURIParam(listOfIDs[0:itemsToFetchWithBuffer])
	url := constructURL(ids, itemsToFetchWithBuffer)

	client := resty.New()
	client.SetTimeout(10 * time.Second)

	response, err := client.R().
		SetHeader("User-Agent", app.Name+"/"+app.Version).
		Get(url)
	if err != nil {
		return nil, err.Error()
	}

	sanitizedResponse := ansi.Strip(string(response.Body()))

	var algoliaItems *endpoints.Algolia
	if err := json.Unmarshal([]byte(sanitizedResponse), &algoliaItems); err != nil {
		return nil, fmt.Sprintf("Error while unmarshalling sanitized response: %v", err)
	}

	mapOfItemsWithMetaData := mapStories(algoliaItems)
	orderedStories := joinStories(listOfIDs, mapOfItemsWithMetaData)

	return orderedStories[0:min(itemsToFetch, len(orderedStories))], ""
}

func constructURL(ids string, count int) string {
	baseURL := "https://hn.algolia.com/api/v1/search?tags=story,"
	return baseURL + "(" + ids + ")&hitsPerPage=" + strconv.Itoa(count)
}

func fetchStoriesList(category int) (stories []int, errMsg string) {
	url := fmt.Sprintf("%s/%s.json", uri, getCategory(category))

	client := resty.New()
	client.SetTimeout(10 * time.Second)

	_, err := client.R().
		SetHeader("User-Agent", app.Name+"/"+app.Version).
		SetResult(&stories).
		Get(url)
	if err != nil {
		return nil, err.Error()
	}

	return stories, ""
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
	case category.Top:
		return "topstories"

	case category.New:
		return "newstories"

	case category.Ask:
		return "askstories"

	case category.Show:
		return "showstories"

	case category.Best:
		return "beststories"

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
			Title:         sanitize(story.Title),
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

func sanitize(s string) string {
	var b strings.Builder

	for _, c := range s {
		if c == '­' {
			continue
		}

		b.WriteRune(c)
	}

	return b.String()
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

func (s *Service) FetchItem(id int) *item.Item {
	hn := new(endpoints.HN)

	client := resty.New()
	client.SetTimeout(5 * time.Second)
	client.SetBaseURL("https://hacker-news.firebaseio.com/v0/item/")

	_, err := client.R().
		SetHeader("User-Agent", app.Name+"/"+app.Version).
		SetResult(hn).
		Get(strconv.Itoa(id) + ".json")
	if err != nil {
		panic(err)
	}

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

func (s *Service) FetchComments(id int) *item.Item {
	client := resty.New()
	client.SetTimeout(5 * time.Second)
	client.SetBaseURL("http://api.hackerwebapp.com/item/")

	response, err := client.R().
		SetHeader("User-Agent", app.Name+"/"+app.Version).
		Get(strconv.Itoa(id))
	if err != nil {
		panic(err)
	}

	sanitizedResponse := ansi.Strip(string(response.Body()))

	comments := new(endpoints.Comments)
	if err := json.Unmarshal([]byte(sanitizedResponse), comments); err != nil {
		panic(fmt.Sprintf("Error while unmarshalling sanitized response: %v", err))
	}

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
