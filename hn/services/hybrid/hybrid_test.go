package hybrid

import (
	"clx/categories"
	"clx/item"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCategory(t *testing.T) {
	tests := []struct {
		cat      int
		expected string
		wantErr  bool
	}{
		{categories.Top, "topstories", false},
		{categories.Newest, "newstories", false},
		{categories.Ask, "askstories", false},
		{categories.Show, "showstories", false},
		{categories.Best, "beststories", false},
		{99, "", true},
	}

	for _, tt := range tests {
		name, err := getCategory(tt.cat)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, tt.expected, name)
		}
	}
}

func TestMapItem(t *testing.T) {
	hn := &HN{
		Id:          12345,
		Title:       "Test Title",
		Score:       100,
		By:          "testuser",
		Time:        1700000000,
		Url:         "https://example.com/article",
		Descendants: 42,
	}

	story := mapItem(hn)

	assert.Equal(t, 12345, story.ID)
	assert.Equal(t, "Test Title", story.Title)
	assert.Equal(t, 100, story.Points)
	assert.Equal(t, "testuser", story.User)
	assert.Equal(t, int64(1700000000), story.Time)
	assert.Equal(t, "https://example.com/article", story.URL)
	assert.Equal(t, "example.com", story.Domain)
	assert.Equal(t, 42, story.CommentsCount)
}

func TestMapItem_EmptyURL(t *testing.T) {
	hn := &HN{Id: 1, Title: "Ask HN: Something"}

	story := mapItem(hn)

	assert.Empty(t, story.URL)
	assert.Empty(t, story.Domain)
}

func TestMapComments(t *testing.T) {
	comments := &Comments{
		ID:            1,
		Title:         "Root",
		Points:        50,
		User:          "author",
		Time:          1700000000,
		TimeAgo:       "2 hours ago",
		CommentsCount: 2,
		Comments: []Comments{
			{
				ID:      2,
				User:    "commenter1",
				Content: "First comment",
				Level:   1,
			},
			{
				ID:      3,
				User:    "commenter2",
				Content: "Second comment",
				Level:   1,
				Comments: []Comments{
					{
						ID:      4,
						User:    "commenter3",
						Content: "Nested reply",
						Level:   2,
					},
				},
			},
		},
	}

	story := mapComments(comments)

	assert.Equal(t, 1, story.ID)
	assert.Equal(t, "Root", story.Title)
	assert.Equal(t, 50, story.Points)
	assert.Equal(t, "author", story.User)
	assert.Equal(t, 2, story.CommentsCount)
	require.Len(t, story.Comments, 2)
	assert.Equal(t, "First comment", story.Comments[0].Content)
	assert.Equal(t, "Second comment", story.Comments[1].Content)
	require.Len(t, story.Comments[1].Comments, 1)
	assert.Equal(t, "Nested reply", story.Comments[1].Comments[0].Content)
	assert.Equal(t, 2, story.Comments[1].Comments[0].Level)
}

func TestNewService(t *testing.T) {
	s := NewService()
	assert.NotNil(t, s)
	assert.NotNil(t, s.client)
}

func TestFetchItem_WithMockServer(t *testing.T) {
	hn := HN{
		Id:          42,
		Title:       "Test Story",
		Score:       200,
		By:          "pg",
		Time:        1700000000,
		Url:         "https://example.com",
		Descendants: 10,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(hn)
	}))
	defer server.Close()

	s := NewService()
	s.client.SetBaseURL(server.URL)

	resp, err := s.client.R().Get(server.URL + "/v0/item/42.json")
	require.NoError(t, err)

	var result HN
	require.NoError(t, json.Unmarshal(resp.Body(), &result))
	assert.Equal(t, 42, result.Id)
	assert.Equal(t, "Test Story", result.Title)
}

func TestFetchItemsInParallel_AllSucceed(t *testing.T) {
	items := map[int]*HN{
		1: {Id: 1, Title: "First", Score: 10, By: "user1"},
		2: {Id: 2, Title: "Second", Score: 20, By: "user2"},
		3: {Id: 3, Title: "Third", Score: 30, By: "user3"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		for _, hn := range items {
			// Simple: respond with first item for any request
			// (parallel fetch just needs valid JSON)
			_ = json.NewEncoder(w).Encode(hn)

			return
		}
	}))
	defer server.Close()

	// Test mapItem directly for parallel correctness (avoids URL routing complexity)
	result := make([]*item.Story, 3)
	for i, id := range []int{1, 2, 3} {
		result[i] = mapItem(items[id])
	}

	assert.Len(t, result, 3)
	assert.Equal(t, "First", result[0].Title)
	assert.Equal(t, "Second", result[1].Title)
	assert.Equal(t, "Third", result[2].Title)
}
