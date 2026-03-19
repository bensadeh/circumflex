package firebase

import (
	"clx/ansi"
	"clx/item"
	"clx/timeago"
	"clx/version"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/go-resty/resty/v2"
)

const (
	defaultBaseURL = "https://hacker-news.firebaseio.com/v0"
	maxConcurrency = 25
)

type Service struct {
	client  *resty.Client
	baseURL string
}

func NewService() *Service {
	client := resty.New()
	client.SetTimeout(10 * time.Second)
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	client.SetHeader("User-Agent", version.Name+"/"+version.Version)

	return &Service{client: client, baseURL: defaultBaseURL}
}

func (s *Service) FetchItems(ctx context.Context, itemsToFetch int, category string) ([]*item.Story, error) {
	ids, err := s.fetchStoriesList(ctx, category)
	if err != nil {
		return nil, err
	}

	ids = ids[:min(len(ids), itemsToFetch)]

	return s.fetchItemsInParallel(ctx, ids)
}

func (s *Service) fetchStoriesList(ctx context.Context, category string) ([]int, error) {
	var ids []int

	url := fmt.Sprintf("%s/%s.json", s.baseURL, category)

	resp, err := s.client.R().SetContext(ctx).SetResult(&ids).Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("could not fetch stories, server returned status %d %s",
			resp.StatusCode(), http.StatusText(resp.StatusCode()))
	}

	return ids, nil
}

func (s *Service) fetchItemsInParallel(ctx context.Context, ids []int) ([]*item.Story, error) {
	items := make([]*item.Story, len(ids))

	var wg sync.WaitGroup

	for i, id := range ids {
		wg.Add(1)

		go func(i, id int) {
			defer wg.Done()

			hn, err := s.fetchHNItem(ctx, id)
			if err == nil {
				items[i] = mapStoryItem(hn)
			}
		}(i, id)
	}

	wg.Wait()

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

func (s *Service) FetchItem(ctx context.Context, id int) (*item.Story, error) {
	hn, err := s.fetchHNItem(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapStoryItem(hn), nil
}

func (s *Service) FetchComments(ctx context.Context, id int) (*item.Story, error) {
	hn, err := s.fetchHNItem(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching story %d: %w", id, err)
	}

	story := mapRootItem(hn)

	sem := make(chan struct{}, maxConcurrency)
	story.Comments = s.fetchCommentTree(ctx, sem, hn.Kids, 0)

	return story, nil
}

func (s *Service) fetchCommentTree(ctx context.Context, sem chan struct{}, kidIDs []int, level int) []*item.Story {
	if len(kidIDs) == 0 {
		return nil
	}

	comments := make([]*item.Story, len(kidIDs))

	var wg sync.WaitGroup

	for i, kidID := range kidIDs {
		wg.Add(1)

		go func(i, kidID int) {
			defer wg.Done()

			// Acquire semaphore, respecting context cancellation.
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}

			hn, err := s.fetchHNItem(ctx, kidID)

			// Release semaphore before recursion to avoid deadlock: child
			// goroutines can acquire slots while the parent continues.
			<-sem

			if err != nil {
				return
			}

			if hn.Dead {
				return
			}

			c := mapCommentItem(hn, level)
			c.Comments = s.fetchCommentTree(ctx, sem, hn.Kids, level+1)
			comments[i] = c
		}(i, kidID)
	}

	wg.Wait()

	return filterNil(comments)
}

func (s *Service) fetchHNItem(ctx context.Context, id int) (*hnItem, error) {
	url := fmt.Sprintf("%s/item/%d.json", s.baseURL, id)

	resp, err := s.client.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching item %d: %w", id, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("could not fetch item %d, server returned status %d %s",
			id, resp.StatusCode(), http.StatusText(resp.StatusCode()))
	}

	// Strip ANSI escape sequences as a defensive measure against potential
	// injection in user-submitted content (title, text fields).
	sanitized := ansi.Strip(string(resp.Body()))

	if sanitized == "null" {
		return nil, fmt.Errorf("item %d does not exist", id)
	}

	var hn hnItem
	if err := json.Unmarshal([]byte(sanitized), &hn); err != nil {
		return nil, fmt.Errorf("unexpected response from server for item %d: %w", id, err)
	}

	return &hn, nil
}

func mapStoryItem(hn *hnItem) *item.Story {
	return &item.Story{
		ID:            hn.ID,
		Title:         hn.Title,
		Points:        hn.Score,
		User:          hn.By,
		Time:          hn.Time,
		URL:           hn.URL,
		Domain:        domainutil.Domain(hn.URL),
		CommentsCount: hn.Descendants,
	}
}

func mapRootItem(hn *hnItem) *item.Story {
	return &item.Story{
		ID:            hn.ID,
		Title:         hn.Title,
		Points:        hn.Score,
		User:          hn.By,
		Time:          hn.Time,
		TimeAgo:       timeago.RelativeTime(hn.Time),
		Type:          hn.Type,
		URL:           hn.URL,
		Domain:        domainutil.Domain(hn.URL),
		Content:       hn.Text,
		CommentsCount: hn.Descendants,
	}
}

func mapCommentItem(hn *hnItem, level int) *item.Story {
	content := hn.Text
	if hn.Deleted {
		content = "[deleted]"
	}

	return &item.Story{
		ID:      hn.ID,
		User:    hn.By,
		Time:    hn.Time,
		TimeAgo: timeago.RelativeTime(hn.Time),
		Type:    hn.Type,
		Level:   level,
		Content: content,
	}
}

func filterNil(items []*item.Story) []*item.Story {
	if len(items) == 0 {
		return nil
	}

	result := make([]*item.Story, 0, len(items))

	for _, it := range items {
		if it != nil {
			result = append(result, it)
		}
	}

	return result
}
