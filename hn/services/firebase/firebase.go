package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/item"
	"github.com/bensadeh/circumflex/timeago"
	"github.com/bensadeh/circumflex/version"

	"github.com/bobesa/go-domain-util/domainutil"
	"resty.dev/v3"
)

const (
	defaultBaseURL = "https://hacker-news.firebaseio.com/v0"
	maxConcurrency = 25
	httpTimeout    = 10 * time.Second
	retryCount     = 3
	retryWaitTime  = 200 * time.Millisecond
	retryMaxWait   = 2 * time.Second
)

var errItemNotFound = errors.New("item not found")

// discardLogger silences resty's internal logging so that WARN/ERROR
// messages on context cancellation don't corrupt the TUI.
type discardLogger struct{}

func (discardLogger) Errorf(string, ...any) {}
func (discardLogger) Warnf(string, ...any)  {}
func (discardLogger) Debugf(string, ...any) {}

type Service struct {
	client  *resty.Client
	baseURL string
}

func NewService() *Service {
	client := resty.New()
	client.SetTimeout(httpTimeout)
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	client.SetHeader("User-Agent", version.Name+"/"+version.Version)
	client.SetRetryCount(retryCount)
	client.SetRetryWaitTime(retryWaitTime)
	client.SetRetryMaxWaitTime(retryMaxWait)
	client.AddRetryConditions(func(resp *resty.Response, _ error) bool {
		return resp != nil && resp.StatusCode() >= http.StatusInternalServerError
	})
	client.SetLogger(discardLogger{})

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

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	var (
		wg       sync.WaitGroup
		firstErr error
		errOnce  sync.Once
	)

	fail := func(err error) {
		errOnce.Do(func() {
			firstErr = err
			cancel(err)
		})
	}

	for i, id := range ids {
		wg.Add(1)

		go func(i, id int) {
			defer wg.Done()

			hn, err := s.fetchHNItem(ctx, id)
			if err != nil {
				if !errors.Is(err, errItemNotFound) {
					fail(err)
				}

				return
			}

			items[i] = mapStoryItem(hn)
		}(i, id)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return filterNil(items), nil
}

func (s *Service) FetchItem(ctx context.Context, id int) (*item.Story, error) {
	hn, err := s.fetchHNItem(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapStoryItem(hn), nil
}

func (s *Service) FetchComments(ctx context.Context, id int, onProgress func(fetched, total int)) (*item.Story, error) {
	hn, err := s.fetchHNItem(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching story %d: %w", id, err)
	}

	story := mapRootItem(hn)

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	sem := make(chan struct{}, maxConcurrency)

	var fetched atomic.Int64

	story.Comments, err = s.fetchCommentTree(ctx, cancel, sem, hn.Kids, &fetched, hn.Descendants, onProgress)
	if err != nil {
		return nil, fmt.Errorf("fetching comments for story %d: %w", id, err)
	}

	return story, nil
}

func (s *Service) fetchCommentTree(ctx context.Context, cancel context.CancelCauseFunc, sem chan struct{}, kidIDs []int, fetched *atomic.Int64, total int, onProgress func(fetched, total int)) ([]*item.Story, error) {
	if len(kidIDs) == 0 {
		return nil, nil
	}

	comments := make([]*item.Story, len(kidIDs))

	var (
		wg       sync.WaitGroup
		firstErr error
		errOnce  sync.Once
	)

	fail := func(err error) {
		errOnce.Do(func() {
			firstErr = err
			cancel(err)
		})
	}

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
				if !errors.Is(err, errItemNotFound) {
					fail(err)
				}

				return
			}

			if onProgress != nil {
				onProgress(int(fetched.Add(1)), total)
			}

			if hn.Dead {
				return
			}

			c := mapCommentItem(hn)

			children, err := s.fetchCommentTree(ctx, cancel, sem, hn.Kids, fetched, total, onProgress)
			if err != nil {
				fail(err)

				return
			}

			c.Comments = children
			comments[i] = c
		}(i, kidID)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return filterNil(comments), nil
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
	sanitized := ansi.Strip(string(resp.Bytes()))

	if sanitized == "null" {
		return nil, fmt.Errorf("item %d: %w", id, errItemNotFound)
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
		URL:           hn.URL,
		Domain:        domainutil.Domain(hn.URL),
		Content:       hn.Text,
		CommentsCount: hn.Descendants,
	}
}

func mapCommentItem(hn *hnItem) *item.Story {
	content := hn.Text
	if hn.Deleted {
		content = "[deleted]"
	}

	return &item.Story{
		ID:      hn.ID,
		User:    hn.By,
		Time:    hn.Time,
		TimeAgo: timeago.RelativeTime(hn.Time),
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
