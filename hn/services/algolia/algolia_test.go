package algolia

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/bensadeh/circumflex/hn"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapHit(t *testing.T) {
	hit := &searchHit{
		ObjectID:    "12345",
		Title:       "Test Title",
		URL:         "https://example.com/article",
		Author:      "testuser",
		Points:      100,
		NumComments: 42,
		CreatedAtI:  1700000000,
	}

	story := mapHit(hit)

	require.NotNil(t, story)
	assert.Equal(t, 12345, story.ID)
	assert.Equal(t, "Test Title", story.Title)
	assert.Equal(t, 100, story.Points)
	assert.Equal(t, "testuser", story.Author)
	assert.Equal(t, int64(1700000000), story.Time)
	assert.Equal(t, "https://example.com/article", story.URL)
	assert.Equal(t, "example.com", story.Domain)
	assert.Equal(t, 42, story.CommentsCount)
}

func TestMapHit_EmptyURL(t *testing.T) {
	hit := &searchHit{ObjectID: "1", Title: "Ask HN: Something"}

	story := mapHit(hit)

	require.NotNil(t, story)
	assert.Empty(t, story.URL)
	assert.Empty(t, story.Domain)
}

func TestMapHit_NonNumericIDSkipped(t *testing.T) {
	hit := &searchHit{ObjectID: "not-a-number", Title: "Broken"}

	assert.Nil(t, mapHit(hit))
}

func TestMapHit_StripsANSIFromUserFields(t *testing.T) {
	hit := &searchHit{
		ObjectID: "7",
		Title:    "Evil \x1b[31mtitle\x1b[0m",
		Author:   "user\x1b[2J",
		URL:      "https://example.com/\x1b]8;;http://evil.com\x1b\\path",
	}

	story := mapHit(hit)

	require.NotNil(t, story)
	assert.Equal(t, "Evil title", story.Title)
	assert.Equal(t, "user", story.Author)
	assert.NotContains(t, story.URL, "\x1b")
}

func TestSearchItems_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/search", r.URL.Path)
		assert.Equal(t, "gpu", r.URL.Query().Get("query"))
		assert.Equal(t, "story", r.URL.Query().Get("tags"))
		assert.Equal(t, "30", r.URL.Query().Get("hitsPerPage"))
		assert.Empty(t, r.URL.Query().Get("numericFilters"), "all time sends no age filter")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"hits": [
				{
					"objectID": "100",
					"title": "First hit",
					"url": "https://example.com/one",
					"author": "alfa",
					"points": 55,
					"num_comments": 7,
					"created_at_i": 1700000000
				},
				{
					"objectID": "200",
					"title": "Self post",
					"author": "beta",
					"points": 12,
					"num_comments": null,
					"created_at_i": 1700000500
				}
			]
		}`))
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	stories, err := s.SearchItems(context.Background(), hn.SearchRequest{Query: "gpu", ItemsToFetch: 30})
	require.NoError(t, err)
	require.Len(t, stories, 2)

	assert.Equal(t, 100, stories[0].ID)
	assert.Equal(t, "First hit", stories[0].Title)
	assert.Equal(t, "example.com", stories[0].Domain)

	assert.Equal(t, 200, stories[1].ID)
	assert.Empty(t, stories[1].URL)
	assert.Zero(t, stories[1].CommentsCount, "null num_comments maps to zero")
}

func TestSearchItems_FiltersMapToParams(t *testing.T) {
	before := time.Now().Add(-7 * 24 * time.Hour).Unix()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/search_by_date", r.URL.Path, "date sort uses the by-date endpoint")
		assert.Equal(t, "story", r.URL.Query().Get("tags"))

		filter := r.URL.Query().Get("numericFilters")
		if assert.Greater(t, len(filter), len("created_at_i>"), "age filter must be sent") {
			oldest, err := strconv.ParseInt(filter[len("created_at_i>"):], 10, 64)
			if assert.NoError(t, err) {
				assert.GreaterOrEqual(t, oldest, before, "cutoff sits a week back from now")
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hits": []}`))
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	_, err := s.SearchItems(context.Background(), hn.SearchRequest{
		Query:        "gpu",
		SortByDate:   true,
		MaxAge:       7 * 24 * time.Hour,
		ItemsToFetch: 30,
	})
	require.NoError(t, err)
}

func TestSearchItems_NoHits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hits": []}`))
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	stories, err := s.SearchItems(context.Background(), hn.SearchRequest{Query: "zzzznope", ItemsToFetch: 30})
	require.NoError(t, err)
	assert.Empty(t, stories)
}

func TestSearchItems_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	_, err := s.SearchItems(context.Background(), hn.SearchRequest{Query: "gpu", ItemsToFetch: 30})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
