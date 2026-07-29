package article

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	nurl "net/url"
	"strings"
	"sync/atomic"
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

// A <base href> declaration moves where relative references point; the first
// one carrying href wins, and anything unusable falls back to the page URL.
func TestDocumentBaseURL(t *testing.T) {
	t.Parallel()

	page, err := nurl.Parse("https://example.com/news/deep/article.html")
	require.NoError(t, err)

	tests := []struct {
		name string
		body string
		want string
	}{
		{"no base tag", "<html><head></head><body></body></html>", "https://example.com/news/deep/article.html"},
		{"absolute href", `<html><head><base href="https://example.com/" /></head></html>`, "https://example.com/"},
		{"relative href resolves against the page", `<head><base href="/assets/"></head>`, "https://example.com/assets/"},
		{"first base with href wins", `<head><base target="_blank"><base href="/a/"><base href="/b/"></head>`, "https://example.com/a/"},
		{"invalid href falls back", `<head><base href="%zz"></head>`, "https://example.com/news/deep/article.html"},
		{"non-web scheme falls back", `<head><base href="data:text/html,x"></head>`, "https://example.com/news/deep/article.html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, documentBaseURL([]byte(tt.body), page).String())
		})
	}
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

	parsed, err := Parse(context.Background(), server.URL+"/article", false)
	require.NoError(t, err)

	assert.Equal(t, "Real Title", parsed.Title)
}

// Only a Kitty-graphics terminal can draw an image, so everywhere else the
// download, decode and re-encode would be spent to render a text label.
func TestParse_SkipsImageFetchWhenUndrawable(t *testing.T) {
	t.Parallel()

	prose := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 20)

	var requested atomic.Int32

	// Encoded once, up front: a handler cannot fail the test from its own
	// goroutine, and re-encoding per request would starve the fetch timeout.
	var photo bytes.Buffer
	require.NoError(t, png.Encode(&photo, image.NewRGBA(image.Rect(0, 0, 100, 100))))

	mux := http.NewServeMux()
	mux.HandleFunc("/article", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body><article><p>" + prose + "</p>" +
			`<p><img src="/photo.png" width="600" alt="a caption"></p>` +
			"<p>" + prose + "</p></article></body></html>"))
	})
	mux.HandleFunc("/photo.png", func(w http.ResponseWriter, _ *http.Request) {
		requested.Add(1)

		_, _ = w.Write(photo.Bytes())
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	parsed, err := Parse(context.Background(), server.URL+"/article", false)
	require.NoError(t, err)
	assert.Zero(t, requested.Load(), "an undrawable image is never downloaded")

	block := firstImageBlock(t, parsed)
	assert.Nil(t, block.kitty)
	assert.Zero(t, block.imgSize)
	assert.NotEmpty(t, block.imageURL, "the source is kept, so the label still names the image")

	parsed, err = Parse(context.Background(), server.URL+"/article", true)
	require.NoError(t, err)
	assert.Equal(t, int32(1), requested.Load(), "a drawable image is fetched as before")

	assert.NotNil(t, firstImageBlock(t, parsed).kitty)
}

// A page declaring <base href> writes its asset paths against it — browsers
// resolve them there, so the image must download from where the base points,
// not from the page's own directory.
func TestParse_HonorsBaseTagForImages(t *testing.T) {
	t.Parallel()

	prose := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 20)

	var right, wrong atomic.Int32

	var photo bytes.Buffer
	require.NoError(t, png.Encode(&photo, image.NewRGBA(image.Rect(0, 0, 100, 100))))

	mux := http.NewServeMux()
	mux.HandleFunc("/nested/article", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><base href="/" /></head><body><article><p>` + prose + "</p>" +
			`<p><img src="assets/photo.png" width="600" alt="a caption"></p>` +
			"<p>" + prose + "</p></article></body></html>"))
	})
	mux.HandleFunc("/assets/photo.png", func(w http.ResponseWriter, _ *http.Request) {
		right.Add(1)

		_, _ = w.Write(photo.Bytes())
	})
	mux.HandleFunc("/nested/assets/photo.png", func(w http.ResponseWriter, r *http.Request) {
		wrong.Add(1)

		http.NotFound(w, r)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	parsed, err := Parse(context.Background(), server.URL+"/nested/article", true)
	require.NoError(t, err)

	assert.Equal(t, int32(1), right.Load(), "the image downloads from where the base points")
	assert.Zero(t, wrong.Load(), "the page's own directory is never tried")

	block := firstImageBlock(t, parsed)
	assert.Equal(t, server.URL+"/assets/photo.png", block.imageURL)
	assert.NotNil(t, block.kitty)
}

func firstImageBlock(t *testing.T, parsed *Parsed) *block {
	t.Helper()

	for i := range parsed.blocks {
		if parsed.blocks[i].kind == blockImage {
			return &parsed.blocks[i]
		}
	}

	require.FailNow(t, "no image block parsed")

	return nil
}
