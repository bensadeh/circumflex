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
