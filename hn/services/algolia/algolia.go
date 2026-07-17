// Package algolia searches Hacker News through the Algolia HN Search API.
// Hits carry the HN item ID as their objectID, so results bridge straight
// into the Firebase pipeline that serves comments and items.
package algolia

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/domain"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/version"

	"resty.dev/v3"
)

const (
	defaultBaseURL = "https://hn.algolia.com/api/v1"
	httpTimeout    = 10 * time.Second
	retryCount     = 3
	retryWaitTime  = 200 * time.Millisecond
	retryMaxWait   = 2 * time.Second
)

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

type searchResponse struct {
	Hits []searchHit `json:"hits"`
}

type searchHit struct {
	ObjectID    string `json:"objectID"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Author      string `json:"author"`
	Points      int    `json:"points"`
	NumComments int    `json:"num_comments"`
	CreatedAtI  int64  `json:"created_at_i"`
}

// SearchItems returns the stories matching the request, in a single fetch —
// search paginates locally like every other category.
func (s *Service) SearchItems(ctx context.Context, req hn.SearchRequest) ([]*hn.Story, error) {
	endpoint := "/search"
	if req.SortByDate {
		endpoint = "/search_by_date"
	}

	params := map[string]string{
		"query":       req.Query,
		"tags":        "story",
		"hitsPerPage": strconv.Itoa(req.ItemsToFetch),
	}

	if req.MaxAge > 0 {
		oldest := time.Now().Add(-req.MaxAge).Unix()
		params["numericFilters"] = "created_at_i>" + strconv.FormatInt(oldest, 10)
	}

	var result searchResponse

	resp, err := s.client.R().
		SetContext(ctx).
		SetQueryParams(params).
		SetResult(&result).
		Get(s.baseURL + endpoint)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("could not search stories, server returned status %d %s",
			resp.StatusCode(), http.StatusText(resp.StatusCode()))
	}

	stories := make([]*hn.Story, 0, len(result.Hits))

	for _, hit := range result.Hits {
		if story := mapHit(&hit); story != nil {
			stories = append(stories, story)
		}
	}

	return stories, nil
}

func mapHit(hit *searchHit) *hn.Story {
	id, err := strconv.Atoi(hit.ObjectID)
	if err != nil {
		return nil
	}

	// Defend against terminal injection via user-submitted fields.
	url := ansi.Strip(hit.URL)

	return &hn.Story{
		ID:            id,
		Title:         ansi.Strip(hit.Title),
		Points:        hit.Points,
		Author:        ansi.Strip(hit.Author),
		Time:          hit.CreatedAtI,
		URL:           url,
		Domain:        domain.FromURL(url),
		CommentsCount: hit.NumComments,
	}
}
