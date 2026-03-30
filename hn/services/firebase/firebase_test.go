package firebase

import (
	"clx/item"
	"clx/timeago"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapStoryItem(t *testing.T) {
	hn := &hnItem{
		ID:          12345,
		Title:       "Test Title",
		Score:       100,
		By:          "testuser",
		Time:        1700000000,
		URL:         "https://example.com/article",
		Descendants: 42,
	}

	story := mapStoryItem(hn)

	assert.Equal(t, 12345, story.ID)
	assert.Equal(t, "Test Title", story.Title)
	assert.Equal(t, 100, story.Points)
	assert.Equal(t, "testuser", story.User)
	assert.Equal(t, int64(1700000000), story.Time)
	assert.Equal(t, "https://example.com/article", story.URL)
	assert.Equal(t, "example.com", story.Domain)
	assert.Equal(t, 42, story.CommentsCount)
	assert.Empty(t, story.TimeAgo)
}

func TestMapStoryItem_EmptyURL(t *testing.T) {
	hn := &hnItem{ID: 1, Title: "Ask HN: Something"}

	story := mapStoryItem(hn)

	assert.Empty(t, story.URL)
	assert.Empty(t, story.Domain)
}

func TestMapRootItem(t *testing.T) {
	hn := &hnItem{
		ID:          12345,
		Title:       "Test Story",
		Score:       200,
		By:          "author",
		Time:        time.Now().Add(-2 * time.Hour).Unix(),
		Type:        "story",
		URL:         "https://example.com",
		Text:        "<p>Self post content",
		Descendants: 15,
	}

	story := mapRootItem(hn)

	assert.Equal(t, 12345, story.ID)
	assert.Equal(t, "Test Story", story.Title)
	assert.Equal(t, 200, story.Points)
	assert.Equal(t, "author", story.User)
	assert.Equal(t, "https://example.com", story.URL)
	assert.Equal(t, "example.com", story.Domain)
	assert.Equal(t, "<p>Self post content", story.Content)
	assert.Equal(t, 15, story.CommentsCount)
	assert.Contains(t, story.TimeAgo, "hours ago")
}

func TestMapCommentItem(t *testing.T) {
	hn := &hnItem{
		ID:   100,
		By:   "commenter",
		Time: time.Now().Add(-30 * time.Minute).Unix(),
		Text: "This is a comment",
		Type: "comment",
	}

	comment := mapCommentItem(hn)

	assert.Equal(t, 100, comment.ID)
	assert.Equal(t, "commenter", comment.User)
	assert.Equal(t, "This is a comment", comment.Content)
	assert.Contains(t, comment.TimeAgo, "minutes ago")
}

func TestMapCommentItem_Deleted(t *testing.T) {
	hn := &hnItem{
		ID:      101,
		Time:    time.Now().Unix(),
		Deleted: true,
		Type:    "comment",
	}

	comment := mapCommentItem(hn)

	assert.Equal(t, "[deleted]", comment.Content)
	assert.Empty(t, comment.User)
}

func TestFilterNil(t *testing.T) {
	items := filterNil(nil)
	assert.Nil(t, items)
}

func TestFilterNil_WithNils(t *testing.T) {
	input := make([]*item.Story, 3)
	input[1] = &item.Story{ID: 5}

	result := filterNil(input)
	require.Len(t, result, 1)
	assert.Equal(t, 5, result[0].ID)
}

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		offset   time.Duration
		contains string
	}{
		{10 * time.Second, "few seconds"},
		{5 * time.Minute, "minutes ago"},
		{2 * time.Hour, "hours ago"},
		{2 * 24 * time.Hour, "days ago"},
		{60 * 24 * time.Hour, "months ago"},
		{400 * 24 * time.Hour, "year ago"},
	}

	for _, tt := range tests {
		unixTime := time.Now().Add(-tt.offset).Unix()
		result := timeago.RelativeTime(unixTime)
		assert.Contains(t, result, tt.contains, "offset: %v", tt.offset)
	}
}

func TestFetchHNItem_WithMockServer(t *testing.T) {
	hn := hnItem{
		ID:    42,
		Title: "Test Story",
		Score: 200,
		By:    "pg",
		Time:  1700000000,
		URL:   "https://example.com",
		Type:  "story",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(hn)
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	result, err := s.fetchHNItem(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, 42, result.ID)
	assert.Equal(t, "Test Story", result.Title)
	assert.Equal(t, "pg", result.By)
}

func TestFetchHNItem_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	_, err := s.fetchHNItem(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestFetchHNItem_NullResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("null"))
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	_, err := s.fetchHNItem(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "item not found")
}

func TestFetchComments_WithMockServer(t *testing.T) {
	now := time.Now()

	items := map[int]hnItem{
		1: {
			ID: 1, Title: "Root Story", Score: 100, By: "author",
			Time: now.Add(-time.Hour).Unix(), Type: "story",
			URL: "https://example.com", Descendants: 3, Kids: []int{10, 20},
		},
		10: {
			ID: 10, By: "user1", Time: now.Add(-30 * time.Minute).Unix(),
			Type: "comment", Text: "First comment", Parent: 1, Kids: []int{11},
		},
		11: {
			ID: 11, By: "user2", Time: now.Add(-15 * time.Minute).Unix(),
			Type: "comment", Text: "Nested reply", Parent: 10,
		},
		20: {
			ID: 20, By: "user3", Time: now.Add(-20 * time.Minute).Unix(),
			Type: "comment", Text: "Second comment", Parent: 1,
		},
	}

	server := newMockServer(items)
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	story, err := s.FetchComments(context.Background(), 1)
	require.NoError(t, err)

	assert.Equal(t, 1, story.ID)
	assert.Equal(t, "Root Story", story.Title)
	assert.Equal(t, 100, story.Points)
	assert.Equal(t, "author", story.User)
	assert.Equal(t, 3, story.CommentsCount)

	require.Len(t, story.Comments, 2)

	assert.Equal(t, "First comment", story.Comments[0].Content)
	assert.Equal(t, "user1", story.Comments[0].User)

	require.Len(t, story.Comments[0].Comments, 1)
	assert.Equal(t, "Nested reply", story.Comments[0].Comments[0].Content)
	assert.Equal(t, "user2", story.Comments[0].Comments[0].User)

	assert.Equal(t, "Second comment", story.Comments[1].Content)
}

func TestFetchComments_DeletedComment(t *testing.T) {
	now := time.Now()

	items := map[int]hnItem{
		1: {
			ID: 1, Title: "Story", By: "author", Type: "story",
			Time: now.Unix(), Kids: []int{10},
		},
		10: {
			ID: 10, Deleted: true, Time: now.Unix(),
			Type: "comment", Parent: 1, Kids: []int{11},
		},
		11: {
			ID: 11, By: "user", Time: now.Unix(),
			Type: "comment", Text: "Reply to deleted", Parent: 10,
		},
	}

	server := newMockServer(items)
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	story, err := s.FetchComments(context.Background(), 1)
	require.NoError(t, err)

	require.Len(t, story.Comments, 1)
	assert.Equal(t, "[deleted]", story.Comments[0].Content)

	require.Len(t, story.Comments[0].Comments, 1)
	assert.Equal(t, "Reply to deleted", story.Comments[0].Comments[0].Content)
}

func TestFetchComments_DeadComment(t *testing.T) {
	now := time.Now()

	items := map[int]hnItem{
		1: {
			ID: 1, Title: "Story", By: "author", Type: "story",
			Time: now.Unix(), Kids: []int{10, 20},
		},
		10: {
			ID: 10, Dead: true, By: "spammer", Time: now.Unix(),
			Type: "comment", Text: "Spam content", Parent: 1,
		},
		20: {
			ID: 20, By: "user", Time: now.Unix(),
			Type: "comment", Text: "Good comment", Parent: 1,
		},
	}

	server := newMockServer(items)
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	story, err := s.FetchComments(context.Background(), 1)
	require.NoError(t, err)

	require.Len(t, story.Comments, 1)
	assert.Equal(t, "Good comment", story.Comments[0].Content)
}

func TestFetchItems_WithMockServer(t *testing.T) {
	storyIDs := []int{1, 2, 3}
	items := map[int]hnItem{
		1: {ID: 1, Title: "First", Score: 10, By: "user1", Time: 1700000000},
		2: {ID: 2, Title: "Second", Score: 20, By: "user2", Time: 1700000000},
		3: {ID: 3, Title: "Third", Score: 30, By: "user3", Time: 1700000000},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasSuffix(r.URL.Path, "topstories.json") {
			_ = json.NewEncoder(w).Encode(storyIDs)

			return
		}

		path := r.URL.Path
		path = strings.TrimPrefix(path, "/item/")
		path = strings.TrimSuffix(path, ".json")

		id, _ := strconv.Atoi(path)

		it, ok := items[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		_ = json.NewEncoder(w).Encode(it)
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	result, err := s.FetchItems(context.Background(), 3, "topstories")
	require.NoError(t, err)

	assert.Len(t, result, 3)
	assert.Equal(t, "First", result[0].Title)
	assert.Equal(t, "Second", result[1].Title)
	assert.Equal(t, "Third", result[2].Title)
}

func TestFetchComments_ItemFetchError(t *testing.T) {
	now := time.Now()

	items := map[int]hnItem{
		1: {
			ID: 1, Title: "Story", By: "author", Type: "story",
			Time: now.Unix(), Kids: []int{10, 20},
		},
		10: {
			ID: 10, By: "user1", Time: now.Unix(),
			Type: "comment", Text: "Good comment", Parent: 1,
		},
		// Item 20 is missing — the mock server will return 404.
	}

	server := newMockServer(items)
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL
	s.client.SetRetryCount(0) // disable retries to speed up test

	_, err := s.FetchComments(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "20")
}

func TestFetchComments_NestedFetchError(t *testing.T) {
	now := time.Now()

	items := map[int]hnItem{
		1: {
			ID: 1, Title: "Story", By: "author", Type: "story",
			Time: now.Unix(), Kids: []int{10},
		},
		10: {
			ID: 10, By: "user1", Time: now.Unix(),
			Type: "comment", Text: "Parent comment", Parent: 1, Kids: []int{11},
		},
		// Item 11 is missing — nested fetch will fail.
	}

	server := newMockServer(items)
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL
	s.client.SetRetryCount(0)

	_, err := s.FetchComments(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "11")
}

func TestFetchComments_NullItemSkipped(t *testing.T) {
	now := time.Now()

	items := map[int]hnItem{
		1: {
			ID: 1, Title: "Story", By: "author", Type: "story",
			Time: now.Unix(), Kids: []int{10, 20},
		},
		10: {
			ID: 10, By: "user1", Time: now.Unix(),
			Type: "comment", Text: "Good comment", Parent: 1,
		},
		// Item 20 exists in the map but the server returns "null" (deleted item).
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		path := r.URL.Path
		path = strings.TrimPrefix(path, "/item/")
		path = strings.TrimSuffix(path, ".json")

		id, err := strconv.Atoi(path)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if id == 20 {
			_, _ = w.Write([]byte("null"))

			return
		}

		it, ok := items[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		_ = json.NewEncoder(w).Encode(it)
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	story, err := s.FetchComments(context.Background(), 1)
	require.NoError(t, err)

	require.Len(t, story.Comments, 1)
	assert.Equal(t, "Good comment", story.Comments[0].Content)
}

func TestFetchItems_NullItemSkipped(t *testing.T) {
	storyIDs := []int{1, 2, 3}
	items := map[int]hnItem{
		1: {ID: 1, Title: "First", Score: 10, By: "user1", Time: 1700000000},
		2: {ID: 2, Title: "Second", Score: 20, By: "user2", Time: 1700000000},
		3: {ID: 3, Title: "Third", Score: 30, By: "user3", Time: 1700000000},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasSuffix(r.URL.Path, "topstories.json") {
			_ = json.NewEncoder(w).Encode(storyIDs)

			return
		}

		path := r.URL.Path
		path = strings.TrimPrefix(path, "/item/")
		path = strings.TrimSuffix(path, ".json")

		id, _ := strconv.Atoi(path)

		// Item 2 returns "null" (deleted story).
		if id == 2 {
			_, _ = w.Write([]byte("null"))

			return
		}

		it, ok := items[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		_ = json.NewEncoder(w).Encode(it)
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	result, err := s.FetchItems(context.Background(), 3, "topstories")
	require.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, "First", result[0].Title)
	assert.Equal(t, "Third", result[1].Title)
}

func TestFetchComments_RetrySuccess(t *testing.T) {
	now := time.Now()

	items := map[int]hnItem{
		1: {
			ID: 1, Title: "Story", By: "author", Type: "story",
			Time: now.Unix(), Kids: []int{10},
		},
		10: {
			ID: 10, By: "user1", Time: now.Unix(),
			Type: "comment", Text: "Recovered comment", Parent: 1,
		},
	}

	var mu sync.Mutex

	attempts := map[int]int{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		path := r.URL.Path
		path = strings.TrimPrefix(path, "/item/")
		path = strings.TrimSuffix(path, ".json")

		id, err := strconv.Atoi(path)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		mu.Lock()
		attempts[id]++
		attempt := attempts[id]
		mu.Unlock()

		// Fail the first request for item 10, succeed on retry.
		if id == 10 && attempt == 1 {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		it, ok := items[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		_ = json.NewEncoder(w).Encode(it)
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	story, err := s.FetchComments(context.Background(), 1)
	require.NoError(t, err)

	require.Len(t, story.Comments, 1)
	assert.Equal(t, "Recovered comment", story.Comments[0].Content)
}

func newMockServer(items map[int]hnItem) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		path := r.URL.Path
		path = strings.TrimPrefix(path, "/item/")
		path = strings.TrimSuffix(path, ".json")

		id, err := strconv.Atoi(path)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		it, ok := items[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		_ = json.NewEncoder(w).Encode(it)
	}))
}
