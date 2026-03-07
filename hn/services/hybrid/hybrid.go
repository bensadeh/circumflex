package hybrid

import (
	"clx/ansi"
	"clx/categories"
	"clx/item"
	"clx/version"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/go-resty/resty/v2"
)

const (
	uri = "https://hacker-news.firebaseio.com/v0"
)

type Service struct {
	client *resty.Client
}

func NewService() *Service {
	client := resty.New()
	client.SetTimeout(10 * time.Second)
	client.SetHeader("User-Agent", version.Name+"/"+version.Version)
	return &Service{client: client}
}

func (s *Service) FetchItems(itemsToFetch int, cat int) ([]*item.Story, error) {
	listOfIDs, err := s.fetchStoriesList(cat)
	if err != nil {
		return nil, err
	}

	listOfIdsToFetch := listOfIDs[:min(len(listOfIDs), itemsToFetch)]

	return s.fetchItemsInParallel(listOfIdsToFetch)
}

func (s *Service) fetchItemsInParallel(ids []int) ([]*item.Story, error) {
	items := make([]*item.Story, len(ids))
	var wg sync.WaitGroup

	for i, id := range ids {
		wg.Add(1)
		go func(i int, id int) {
			defer wg.Done()
			fetched, err := s.fetchItem(id)
			if err == nil {
				items[i] = fetched
			}
		}(i, id)
	}

	wg.Wait()

	// Filter out nil items (failed fetches)
	var failed int
	result := make([]*item.Story, 0, len(items))
	for _, it := range items {
		if it != nil {
			result = append(result, it)
		} else {
			failed++
		}
	}

	if failed > 0 {
		return result, fmt.Errorf("could not fetch %d/%d items", failed, len(ids))
	}

	return result, nil
}

func (s *Service) fetchStoriesList(cat int) (stories []int, err error) {
	catName, err := getCategory(cat)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s.json", uri, catName)

	client := s.client
	if client == nil {
		client = resty.New()
		client.SetTimeout(10 * time.Second)
	}

	_, err = client.R().
		SetHeader("User-Agent", version.Name+"/"+version.Version).
		SetResult(&stories).
		Get(url)
	if err != nil {
		return nil, err
	}

	return stories, nil
}

func getCategory(cat int) (string, error) {
	switch cat {
	case categories.Top:
		return "topstories", nil

	case categories.Newest:
		return "newstories", nil

	case categories.Ask:
		return "askstories", nil

	case categories.Show:
		return "showstories", nil

	case categories.Best:
		return "beststories", nil

	default:
		return "", fmt.Errorf("unsupported category: %d", cat)
	}
}

func (s *Service) FetchItem(id int) (*item.Story, error) {
	return s.fetchItem(id)
}

func (s *Service) fetchItem(id int) (*item.Story, error) {
	hn := new(HN)

	client := s.client
	if client == nil {
		client = resty.New()
		client.SetTimeout(10 * time.Second)
	}

	resp, err := client.R().
		SetHeader("User-Agent", version.Name+"/"+version.Version).
		Get("https://hacker-news.firebaseio.com/v0/item/" + strconv.Itoa(id) + ".json")
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

func mapItem(hn *HN) *item.Story {
	return &item.Story{
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

func (s *Service) FetchComments(id int) (*item.Story, error) {
	client := s.client
	if client == nil {
		client = resty.New()
		client.SetTimeout(10 * time.Second)
	}

	response, err := client.R().
		SetHeader("User-Agent", version.Name+"/"+version.Version).
		Get("http://api.hackerwebapp.com/item/" + strconv.Itoa(id))
	if err != nil {
		return nil, fmt.Errorf("fetching comments for %d: %w", id, err)
	}

	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("fetching comments for %d: status %d", id, response.StatusCode())
	}

	sanitizedResponse := ansi.Strip(string(response.Body()))

	comments := new(Comments)
	if err := json.Unmarshal([]byte(sanitizedResponse), comments); err != nil {
		return nil, fmt.Errorf("parsing comments for %d: %w", id, err)
	}

	return mapComments(comments), nil
}

func mapComments(comments *Comments) *item.Story {
	items := make([]*item.Story, 0, len(comments.Comments))

	for i := range comments.Comments {
		items = append(items, mapComments(&comments.Comments[i]))
	}

	return &item.Story{
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
