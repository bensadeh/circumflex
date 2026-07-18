package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
	assert.Equal(t, raw.Time, tree.Time)
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
	assert.Equal(t, raw.Time, node.Time)
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

func TestFetchHNItem_PreservesEscapedBackslashes(t *testing.T) {
	// Regression for #201: `\func` in text used to break JSON parsing.
	raw := hnItem{
		ID:   48264635,
		By:   "user",
		Text: `All commands have the format \func inputs. \begin and \\path also.`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(raw)
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	result, err := s.fetchHNItem(context.Background(), 48264635)
	require.NoError(t, err)
	assert.Equal(t, raw.Text, result.Text)
}

func TestFetchHNItem_StripsANSIFromUserFields(t *testing.T) {
	raw := hnItem{
		ID:    1,
		By:    "\x1B[31muser\x1B[0m",
		Title: "title \x1B[1mbold\x1B[0m end",
		Text:  "body \x07 with bel",
		URL:   "https://example.com/\x1B[2Apath",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(raw)
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	result, err := s.fetchHNItem(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "user", result.By)
	assert.Equal(t, "title bold end", result.Title)
	assert.Equal(t, "body  with bel", result.Text)
	assert.Equal(t, "https://example.com/path", result.URL)
}

func TestFetchHNItem_UnescapesTitleEntities(t *testing.T) {
	raw := hnItem{
		ID:    1,
		Title: "AI Q&amp;A: What&#x27;s new? &#x1b;[31mred",
		Text:  "a &gt; b",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(raw)
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	result, err := s.fetchHNItem(context.Background(), 1)
	require.NoError(t, err)

	// Entities decode, and an entity-encoded escape sequence is stripped
	// rather than reaching the terminal.
	assert.Equal(t, "AI Q&A: What's new? red", result.Title)

	// Text keeps its entities: the comment parser decodes them.
	assert.Equal(t, "a &gt; b", result.Text)
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

func TestFetchItems_BoundedConcurrency(t *testing.T) {
	t.Parallel()

	const itemCount = maxConcurrency + 10

	var inFlight, peak atomic.Int64

	storyIDs := make([]int, itemCount)
	for i := range storyIDs {
		storyIDs[i] = i + 1
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasSuffix(r.URL.Path, "topstories.json") {
			_ = json.NewEncoder(w).Encode(storyIDs)

			return
		}

		current := inFlight.Add(1)
		defer inFlight.Add(-1)

		for {
			observed := peak.Load()
			if current <= observed || peak.CompareAndSwap(observed, current) {
				break
			}
		}

		time.Sleep(10 * time.Millisecond)

		_ = json.NewEncoder(w).Encode(hnItem{ID: 1, Title: "Story", By: "user", Time: 1700000000})
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	result, err := s.FetchItems(context.Background(), itemCount, "topstories")
	require.NoError(t, err)

	assert.Len(t, result, itemCount)
	assert.LessOrEqual(t, peak.Load(), int64(maxConcurrency))
	assert.Greater(t, peak.Load(), int64(1), "requests should actually overlap")
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

func TestFetchCommentNodes_CancelledWhileParked_ReturnsCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("fetch aborted")

	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(cause)

	// A full semaphore parks every goroutine at acquire, so all of them take
	// the silent cancellation exit.
	sem := make(chan struct{}, maxConcurrency)
	for range maxConcurrency {
		sem <- struct{}{}
	}

	var fetched atomic.Int64

	nodes, err := (&Service{}).fetchCommentNodes(ctx, cancel, sem, []int{10, 20}, &fetched, 2, nil)

	require.ErrorIs(t, err, cause)
	assert.Nil(t, nodes)
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

func TestFetchItems_SkipsDeletedAndDead(t *testing.T) {
	storyIDs := []int{1, 2, 3}
	items := map[int]hnItem{
		1: {ID: 1, Title: "First", Score: 10, By: "user1", Time: 1700000000},
		2: {ID: 2, Deleted: true, Time: 1700000000},
		3: {ID: 3, Title: "Third", Score: 30, By: "user3", Time: 1700000000, Dead: true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasSuffix(r.URL.Path, "topstories.json") {
			_ = json.NewEncoder(w).Encode(storyIDs)

			return
		}

		path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/item/"), ".json")
		id, _ := strconv.Atoi(path)
		_ = json.NewEncoder(w).Encode(items[id])
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	result, err := s.FetchItems(context.Background(), 3, "topstories")
	require.NoError(t, err)

	require.Len(t, result, 1)
	assert.Equal(t, "First", result[0].Title)
}

func TestFetchItem_DeletedOrDead(t *testing.T) {
	items := map[int]hnItem{
		2: {ID: 2, Deleted: true},
		3: {ID: 3, Title: "Flagged", Dead: true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/item/"), ".json")
		id, _ := strconv.Atoi(path)
		_ = json.NewEncoder(w).Encode(items[id])
	}))
	defer server.Close()

	s := NewService()
	s.baseURL = server.URL

	_, err := s.FetchItem(context.Background(), 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleted")

	_, err = s.FetchItem(context.Background(), 3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "flagged")
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
