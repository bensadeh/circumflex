package website

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// activePage mirrors the shape of https://news.ycombinator.com/active: two
// submissions, each trailed by the subtext and spacer rows that carry no ID,
// and the page-spacing row that carries one but is not a story.
const activePage = `<html lang="en" op="active"><body><center><table id="hnmain">
<tr id="pagespace" title="Active Threads" style="height:10px"></tr>
<tr class="athing submission" id="49717558">
  <td class="title"><span class="rank">1.</span></td>
  <td class="title"><span class="titleline"><a href="https://example.com/one">One</a></span></td>
</tr>
<tr><td class="subtext"><span class="score" id="score_49717558">925 points</span> by
  <a href="user?id=alfa" class="hnuser">alfa</a>
  <span class="age" title="2026-09-15T19:25:03"><a href="item?id=49717558">8 hours ago</a></span> |
  <a href="item?id=49717558">294&nbsp;comments</a></td></tr>
<tr class="spacer" style="height:5px"></tr>
<tr class="athing submission" id="49708431">
  <td class="title"><span class="titleline"><a href="https://example.com/two">Two</a></span></td>
</tr>
<tr class="spacer" style="height:5px"></tr>
</table></center></body></html>`

func newTestService(t *testing.T, handler http.HandlerFunc) *Service {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	service := NewService()
	service.baseURL = server.URL

	return service
}

func respondWith(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestFetchActiveStoryIDs(t *testing.T) {
	var requestedPath string

	s := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		respondWith(http.StatusOK, activePage)(w, r)
	})

	ids, err := s.FetchActiveStoryIDs(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []int{49717558, 49708431}, ids)
	assert.Equal(t, activePath, requestedPath)
}

// Hacker News answers 403 when it throttles a client rather than failing the
// request outright, so the status has to surface as an error.
func TestFetchActiveStoryIDs_ServerError(t *testing.T) {
	s := newTestService(t, respondWith(http.StatusForbidden, ""))

	ids, err := s.FetchActiveStoryIDs(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
	assert.Nil(t, ids)
}

// A page without submissions means the markup moved, not that Hacker News
// went quiet, so it is an error rather than an empty category.
func TestFetchActiveStoryIDs_NoSubmissions(t *testing.T) {
	s := newTestService(t, respondWith(http.StatusOK, `<table><tr class="spacer"></tr></table>`))

	ids, err := s.FetchActiveStoryIDs(context.Background())

	require.ErrorIs(t, err, errNoStories)
	assert.Nil(t, ids)
}

func TestFetchActiveStoryIDs_ContextCancelled(t *testing.T) {
	s := newTestService(t, func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	ids, err := s.FetchActiveStoryIDs(ctx)

	require.Error(t, err)
	assert.Nil(t, ids)
}

func TestParseStoryIDs(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []int
	}{
		{
			name: "rows keep the order the page ranks them in",
			body: activePage,
			want: []int{49717558, 49708431},
		},
		{
			name: "attribute order does not matter",
			body: `<tr id="42" class="athing submission"></tr>`,
			want: []int{42},
		},
		{
			name: "rows carrying an id but no athing class are skipped",
			body: `<tr id="pagespace"></tr><tr id="7"></tr>`,
			want: nil,
		},
		{
			name: "athing rows without a numeric id are skipped",
			body: `<tr class="athing submission" id="up_42"></tr><tr class="athing"></tr>`,
			want: nil,
		},
		{
			name: "empty page",
			body: "",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := parseStoryIDs(strings.NewReader(tt.body))

			require.NoError(t, err)
			assert.Equal(t, tt.want, ids)
		})
	}
}
