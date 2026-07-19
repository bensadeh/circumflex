package article

import (
	"context"
	"net/http"
	"net/http/httptest"
	nurl "net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A page that redirects must report the URL it landed on: relative
// references, site rules and the image Referer resolve against it.
func TestFetchArticle_ReturnsRedirectTarget(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/moved/article", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/moved/article", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body><p>hello</p></body></html>"))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	url := server.URL + "/start"
	parsed, err := nurl.ParseRequestURI(url)
	require.NoError(t, err)

	_, _, finalURL, err := fetchArticle(context.Background(), url, parsed)
	require.NoError(t, err)

	assert.Equal(t, "/moved/article", finalURL.Path)
}

// The page's own title heads the reader view when a link is followed; escape
// bytes the page smuggles into it — raw or entity-encoded — must not survive
// to the terminal.
func TestParse_StripsEscapesFromPageTitle(t *testing.T) {
	t.Parallel()

	prose := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 20)

	mux := http.NewServeMux()
	mux.HandleFunc("/article", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><head><title>&#27;]0;pwned&#7;Real Title</title></head>" +
			"<body><article><p>" + prose + "</p><p>" + prose + "</p></article></body></html>"))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	parsed, err := Parse(context.Background(), server.URL+"/article")
	require.NoError(t, err)

	assert.Equal(t, "Real Title", parsed.Title)
}
