package firebase

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bensadeh/circumflex/timeago"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapStoryItem(t *testing.T) {
	raw := &hnItem{
		ID:          12345,
		Title:       "Test Title",
		Score:       100,
		By:          "testuser",
		Time:        1700000000,
		URL:         "https://example.com/article",
		Descendants: 42,
	}

	story := mapStoryItem(raw)

	assert.Equal(t, 12345, story.ID)
	assert.Equal(t, "Test Title", story.Title)
	assert.Equal(t, 100, story.Points)
	assert.Equal(t, "testuser", story.Author)
	assert.Equal(t, int64(1700000000), story.Time)
	assert.Equal(t, "https://example.com/article", story.URL)
	assert.Equal(t, "example.com", story.Domain)
	assert.Equal(t, 42, story.CommentsCount)
}

func TestMapStoryItem_EmptyURL(t *testing.T) {
	raw := &hnItem{ID: 1, Title: "Ask HN: Something"}

	story := mapStoryItem(raw)

	assert.Empty(t, story.URL)
	assert.Empty(t, story.Domain)
}

func TestMapCommentTree(t *testing.T) {
	raw := &hnItem{
		ID:          12345,
		Title:       "Test Story",
		Score:       200,
		By:          "author",
		Time:        time.Now().Add(-2 * time.Hour).Unix(),
		URL:         "https://example.com",
		Text:        "<p>Self post content",
		Descendants: 15,
	}

	tree := mapCommentTree(raw)

	assert.Equal(t, 12345, tree.ID)
	assert.Equal(t, "Test Story", tree.Title)
	assert.Equal(t, 200, tree.Points)
	assert.Equal(t, "author", tree.Author)
	assert.Equal(t, "https://example.com", tree.URL)
	assert.Equal(t, "example.com", tree.Domain)
	assert.Equal(t, "<p>Self post content", tree.Content)
	assert.Equal(t, 15, tree.CommentsCount)
	assert.Contains(t, tree.TimeAgo, "hours ago")
}

func TestMapCommentNode(t *testing.T) {
	raw := &hnItem{
		ID:   100,
		By:   "commenter",
		Time: time.Now().Add(-30 * time.Minute).Unix(),
		Text: "This is a comment",
	}

	node := mapCommentNode(raw)

	assert.Equal(t, 100, node.ID)
	assert.Equal(t, "commenter", node.Author)
	assert.Equal(t, "This is a comment", node.Content)
	assert.Contains(t, node.TimeAgo, "minutes ago")
}

func TestMapCommentNode_Deleted(t *testing.T) {
	raw := &hnItem{
		ID:      101,
		Time:    time.Now().Unix(),
		Deleted: true,
	}

	node := mapCommentNode(raw)

	assert.Equal(t, "[deleted]", node.Content)
	assert.Empty(t, node.Author)
}

func TestFilterNil(t *testing.T) {
	result := filterNil[int](nil)
	assert.Nil(t, result)
}

func TestFilterNil_WithNils(t *testing.T) {
	input := make([]*int, 3)
	v := 5
	input[1] = &v

	result := filterNil(input)
	require.Len(t, result, 1)
	assert.Equal(t, 5, *result[0])
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
	raw := hnItem{
		ID:    42,
		Title: "Test Story",
		Score: 200,
		By:    "pg",
		Time:  1700000000,
		URL:   "https://example.com",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(raw)
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
			Time: now.Add(-time.Hour).Unix(),
			URL:  "https://example.com", Descendants: 3, Kids: []int{10, 20},
		},
		10: {
			ID: 10, By: "user1", Time: now.Add(-30 * time.Minute).Unix(),
			Text: "First comment", Kids: []int{11},
		},
		11: {
			ID: 11, By: "user2", Time: now.Add(-15 * time.Minute).Unix(),
			Text: "Nested reply",
		},
		20: {
			ID: 20, By: "user3", Time: now.Add(-20 * time.Minute).Unix(),
			Text: "Second comment",
		},
	}

	server := newMockServer(items)
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	tree, err := s.FetchComments(context.Background(), 1, nil)
	require.NoError(t, err)

	assert.Equal(t, 1, tree.ID)
	assert.Equal(t, "Root Story", tree.Title)
	assert.Equal(t, 100, tree.Points)
	assert.Equal(t, "author", tree.Author)
	assert.Equal(t, 3, tree.CommentsCount)

	require.Len(t, tree.Comments, 2)

	assert.Equal(t, "First comment", tree.Comments[0].Content)
	assert.Equal(t, "user1", tree.Comments[0].Author)

	require.Len(t, tree.Comments[0].Children, 1)
	assert.Equal(t, "Nested reply", tree.Comments[0].Children[0].Content)
	assert.Equal(t, "user2", tree.Comments[0].Children[0].Author)

	assert.Equal(t, "Second comment", tree.Comments[1].Content)
}

func TestFetchComments_DeletedComment(t *testing.T) {
	now := time.Now()

	items := map[int]hnItem{
		1: {
			ID: 1, Title: "Story", By: "author",
			Time: now.Unix(), Kids: []int{10},
		},
		10: {
			ID: 10, Deleted: true, Time: now.Unix(),
			Kids: []int{11},
		},
		11: {
			ID: 11, By: "user", Time: now.Unix(),
			Text: "Reply to deleted",
		},
	}

	server := newMockServer(items)
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	tree, err := s.FetchComments(context.Background(), 1, nil)
	require.NoError(t, err)

	require.Len(t, tree.Comments, 1)
	assert.Equal(t, "[deleted]", tree.Comments[0].Content)

	require.Len(t, tree.Comments[0].Children, 1)
	assert.Equal(t, "Reply to deleted", tree.Comments[0].Children[0].Content)
}

func TestFetchComments_DeadComment(t *testing.T) {
	now := time.Now()

	items := map[int]hnItem{
		1: {
			ID: 1, Title: "Story", By: "author",
			Time: now.Unix(), Kids: []int{10, 20},
		},
		10: {
			ID: 10, Dead: true, By: "spammer", Time: now.Unix(),
			Text: "Spam content",
		},
		20: {
			ID: 20, By: "user", Time: now.Unix(),
			Text: "Good comment",
		},
	}

	server := newMockServer(items)
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	tree, err := s.FetchComments(context.Background(), 1, nil)
	require.NoError(t, err)

	require.Len(t, tree.Comments, 1)
	assert.Equal(t, "Good comment", tree.Comments[0].Content)
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
			ID: 1, Title: "Story", By: "author",
			Time: now.Unix(), Kids: []int{10, 20},
		},
		10: {
			ID: 10, By: "user1", Time: now.Unix(),
			Text: "Good comment",
		},
		// Item 20 is missing — the mock server will return 404.
	}

	server := newMockServer(items)
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL
	s.client.SetRetryCount(0) // disable retries to speed up test

	_, err := s.FetchComments(context.Background(), 1, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "20")
}

func TestFetchComments_NestedFetchError(t *testing.T) {
	now := time.Now()

	items := map[int]hnItem{
		1: {
			ID: 1, Title: "Story", By: "author",
			Time: now.Unix(), Kids: []int{10},
		},
		10: {
			ID: 10, By: "user1", Time: now.Unix(),
			Text: "Parent comment", Kids: []int{11},
		},
		// Item 11 is missing — nested fetch will fail.
	}

	server := newMockServer(items)
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL
	s.client.SetRetryCount(0)

	_, err := s.FetchComments(context.Background(), 1, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "11")
}

func TestFetchComments_NullItemSkipped(t *testing.T) {
	now := time.Now()

	items := map[int]hnItem{
		1: {
			ID: 1, Title: "Story", By: "author",
			Time: now.Unix(), Kids: []int{10, 20},
		},
		10: {
			ID: 10, By: "user1", Time: now.Unix(),
			Text: "Good comment",
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

	tree, err := s.FetchComments(context.Background(), 1, nil)
	require.NoError(t, err)

	require.Len(t, tree.Comments, 1)
	assert.Equal(t, "Good comment", tree.Comments[0].Content)
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
			ID: 1, Title: "Story", By: "author",
			Time: now.Unix(), Kids: []int{10},
		},
		10: {
			ID: 10, By: "user1", Time: now.Unix(),
			Text: "Recovered comment",
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

	tree, err := s.FetchComments(context.Background(), 1, nil)
	require.NoError(t, err)

	require.Len(t, tree.Comments, 1)
	assert.Equal(t, "Recovered comment", tree.Comments[0].Content)
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
