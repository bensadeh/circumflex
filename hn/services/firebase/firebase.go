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
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/version"
	"golang.org/x/sync/errgroup"

	"github.com/bensadeh/circumflex/domain"
	"resty.dev/v3"
)

const (
	defaultBaseURL = "https://hacker-news.firebaseio.com/v0"
	maxConcurrency = 50
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
	client.SetRedirectPolicy(resty.RedirectNoPolicy())
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

func (s *Service) FetchItems(ctx context.Context, itemsToFetch int, category string) ([]*hn.Story, error) {
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

func (s *Service) fetchItemsInParallel(ctx context.Context, ids []int) ([]*hn.Story, error) {
	items := make([]*hn.Story, len(ids))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrency)

	for i, id := range ids {
		g.Go(func() error {
			raw, err := s.fetchHNItem(ctx, id)
			if err != nil {
				// Deleted items are skipped rather than failing the whole page.
				if errors.Is(err, errItemNotFound) {
					return nil
				}

				return err
			}

			items[i] = mapStoryItem(raw)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return filterNil(items), nil
}

func (s *Service) FetchItem(ctx context.Context, id int) (*hn.Story, error) {
	raw, err := s.fetchHNItem(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapStoryItem(raw), nil
}

func (s *Service) FetchComments(ctx context.Context, id int, onProgress func(fetched, total int)) (*hn.CommentTree, error) {
	raw, err := s.fetchHNItem(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching story %d: %w", id, err)
	}

	tree := mapCommentTree(raw)

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	sem := make(chan struct{}, maxConcurrency)

	var fetched atomic.Int64

	tree.Comments, err = s.fetchCommentNodes(ctx, cancel, sem, raw.Kids, &fetched, raw.Descendants, onProgress)
	if err != nil {
		return nil, fmt.Errorf("fetching comments for story %d: %w", id, err)
	}

	return tree, nil
}

func (s *Service) fetchCommentNodes(ctx context.Context, cancel context.CancelCauseFunc, sem chan struct{}, kidIDs []int, fetched *atomic.Int64, total int, onProgress func(fetched, total int)) ([]*hn.CommentNode, error) {
	if len(kidIDs) == 0 {
		return nil, nil
	}

	comments := make([]*hn.CommentNode, len(kidIDs))

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

			raw, err := s.fetchHNItem(ctx, kidID)

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

			if raw.Dead {
				return
			}

			node := mapCommentNode(raw)

			children, err := s.fetchCommentNodes(ctx, cancel, sem, raw.Kids, fetched, total, onProgress)
			if err != nil {
				fail(err)

				return
			}

			node.Children = children
			comments[i] = node
		}(i, kidID)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	// Goroutines parked at the semaphore exit silently on cancellation;
	// without this check a cancelled fetch returns a partial tree with no error.
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}

	return filterNil(comments), nil
}

func (s *Service) fetchHNItem(ctx context.Context, id int) (*hnItem, error) {
	url := fmt.Sprintf("%s/item/%d.json", s.baseURL, id)

	resp, err := s.client.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%s returned status %d %s",
			url, resp.StatusCode(), http.StatusText(resp.StatusCode()))
	}

	body := resp.Bytes()

	if string(body) == "null" {
		return nil, fmt.Errorf("%s: %w", url, errItemNotFound)
	}

	var item hnItem
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("unexpected response from %s: %w", url, err)
	}

	// Defend against terminal injection via user-submitted fields.
	item.By = ansi.Strip(item.By)
	item.Title = ansi.Strip(item.Title)
	item.Text = ansi.Strip(item.Text)
	item.URL = ansi.Strip(item.URL)

	return &item, nil
}

func mapStoryItem(raw *hnItem) *hn.Story {
	return &hn.Story{
		ID:            raw.ID,
		Title:         raw.Title,
		Points:        raw.Score,
		Author:        raw.By,
		Time:          raw.Time,
		URL:           raw.URL,
		Domain:        domain.FromURL(raw.URL),
		CommentsCount: raw.Descendants,
	}
}

func mapCommentTree(raw *hnItem) *hn.CommentTree {
	return &hn.CommentTree{
		Story:   *mapStoryItem(raw),
		Content: raw.Text,
	}
}

func mapCommentNode(raw *hnItem) *hn.CommentNode {
	content := raw.Text
	if raw.Deleted {
		content = "[deleted]"
	}

	return &hn.CommentNode{
		ID:      raw.ID,
		Author:  raw.By,
		Time:    raw.Time,
		Content: content,
	}
}

func filterNil[T any](items []*T) []*T {
	if len(items) == 0 {
		return nil
	}

	result := make([]*T, 0, len(items))

	for _, it := range items {
		if it != nil {
			result = append(result, it)
		}
	}

	return result
}
