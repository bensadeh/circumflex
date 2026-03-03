package hybrid

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/bobesa/go-domain-util/domainutil"

	ansi "clx/utils/strip-ansi"

	"clx/app"
	"clx/constants/category"
	"clx/endpoints"
	"clx/item"

	"github.com/go-resty/resty/v2"
)

const (
	uri = "https://hacker-news.firebaseio.com/v0"
)

type Service struct{}

func (s *Service) FetchItems(itemsToFetch int, category int) (items []*item.Item, errMsg string) {
	listOfIDs, errMsg := fetchStoriesList(category)
	if errMsg != "" {
		return nil, errMsg
	}

	listOfIdsToFetch := listOfIDs[:min(len(listOfIDs), itemsToFetch)]

	return fetchItemsInParallel(listOfIdsToFetch)
}

func fetchItemsInParallel(ids []int) ([]*item.Item, string) {
	items := make([]*item.Item, len(ids))
	var wg sync.WaitGroup

	for i, id := range ids {
		wg.Add(1)
		go func(i int, id int) {
			defer wg.Done()
			fetched, err := fetchItem(id)
			if err == nil {
				items[i] = fetched
			}
		}(i, id)
	}

	wg.Wait()

	// Filter out nil items (failed fetches)
	var failed int
	result := make([]*item.Item, 0, len(items))
	for _, it := range items {
		if it != nil {
			result = append(result, it)
		} else {
			failed++
		}
	}

	var errMsg string
	if failed > 0 {
		errMsg = fmt.Sprintf("Could not fetch %d/%d items", failed, len(ids))
	}

	return result, errMsg
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

func (s *Service) FetchItem(id int) (*item.Item, error) {
	return fetchItem(id)
}

func fetchItem(id int) (*item.Item, error) {
	hn := new(endpoints.HN)

	client := resty.New()
	client.SetTimeout(10 * time.Second)
	client.SetBaseURL("https://hacker-news.firebaseio.com/v0/item/")

	resp, err := client.R().
		SetHeader("User-Agent", app.Name+"/"+app.Version).
		Get(strconv.Itoa(id) + ".json")
	if err != nil {
		return nil, fmt.Errorf("fetching item %d: %w", id, err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("fetching item %d: status %d", id, resp.StatusCode())
	}

	sanitizedBody := ansi.Strip(string(resp.Body()))

	err = json.Unmarshal([]byte(sanitizedBody), hn)
	if err != nil {
		return nil, fmt.Errorf("parsing item %d: %w", id, err)
	}

	return mapItem(hn), nil
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
		Domain:        domainutil.Domain(hn.Url),
		CommentsCount: hn.Descendants,
	}
}

func (s *Service) FetchComments(id int) (*item.Item, error) {
	client := resty.New()
	client.SetTimeout(10 * time.Second)
	client.SetBaseURL("http://api.hackerwebapp.com/item/")

	response, err := client.R().
		SetHeader("User-Agent", app.Name+"/"+app.Version).
		Get(strconv.Itoa(id))
	if err != nil {
		return nil, fmt.Errorf("fetching comments for %d: %w", id, err)
	}

	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("fetching comments for %d: status %d", id, response.StatusCode())
	}

	sanitizedResponse := ansi.Strip(string(response.Body()))

	comments := new(endpoints.Comments)
	if err := json.Unmarshal([]byte(sanitizedResponse), comments); err != nil {
		return nil, fmt.Errorf("parsing comments for %d: %w", id, err)
	}

	return mapComments(comments), nil
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
