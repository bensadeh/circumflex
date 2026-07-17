package article

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	nurl "net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchImages_TinyImageIsDecorativeNotFailed(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size := 100
		if r.URL.Path == "/divider.png" {
			size = 10
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, size, size))); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		_, _ = w.Write(buf.Bytes())
	}))
	defer srv.Close()

	base, err := nurl.Parse(srv.URL)
	require.NoError(t, err)

	blocks := []block{
		{kind: blockImage, imageURL: srv.URL + "/divider.png"},
		{kind: blockImage, imageURL: srv.URL + "/photo.png"},
	}

	fetchImages(context.Background(), blocks, base)

	assert.Nil(t, blocks[0].img)
	assert.True(t, blocks[0].decorative, "below minImageDimension marks the block decorative")

	assert.NotNil(t, blocks[1].img)
	assert.False(t, blocks[1].decorative)
}

func TestFetchImages_SkipsKnownFigures(t *testing.T) {
	t.Parallel()

	requested := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = true

		http.NotFound(w, r)
	}))
	defer srv.Close()

	base, err := nurl.Parse(srv.URL)
	require.NoError(t, err)

	blocks := []block{
		{kind: blockImage, imageURL: srv.URL + "/chart.png", figure: true},
	}

	fetchImages(context.Background(), blocks, base)

	assert.False(t, requested, "figures render their description, never art")
	assert.Nil(t, blocks[0].img)
}

func TestFetchImages_SendsRefererForHotlinkProtection(t *testing.T) {
	t.Parallel()

	var srv *httptest.Server

	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Referer") != srv.URL+"/article" {
			http.Error(w, "hotlinking forbidden", http.StatusForbidden)

			return
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 100, 100))); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		_, _ = w.Write(buf.Bytes())
	}))
	defer srv.Close()

	base, err := nurl.Parse(srv.URL + "/article")
	require.NoError(t, err)

	blocks := []block{{kind: blockImage, imageURL: "photo.png"}}

	fetchImages(context.Background(), blocks, base)

	assert.NotNil(t, blocks[0].img, "same-origin image requests carry the page URL as Referer")
}

func TestRefererFor(t *testing.T) {
	t.Parallel()

	page, err := nurl.Parse("https://example.com/story/index.html#top")
	require.NoError(t, err)

	sameOrigin, err := nurl.Parse("https://example.com/story/photo.webp")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/story/index.html", refererFor(page, sameOrigin),
		"full page URL without fragment for same-origin")

	crossOrigin, err := nurl.Parse("https://cdn.example.net/photo.webp")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/", refererFor(page, crossOrigin),
		"bare origin for cross-origin")

	downgrade, err := nurl.Parse("http://example.com/photo.webp")
	require.NoError(t, err)
	assert.Empty(t, refererFor(page, downgrade), "no referer on https→http downgrade")
}

func TestHasImages(t *testing.T) {
	t.Parallel()

	decoded := &Parsed{blocks: []block{
		{kind: blockParagraph},
		{kind: blockImage, img: image.NewRGBA(image.Rect(0, 0, 100, 100))},
	}}
	assert.True(t, decoded.HasImages())

	undecoded := &Parsed{blocks: []block{{kind: blockImage, imageURL: "https://example.com/a.png"}}}
	assert.False(t, undecoded.HasImages(), "a failed fetch leaves nothing to toggle")

	assert.False(t, NewParsedFromHTML("<p>text only</p>").HasImages())
}

func TestFetchImages_SVGFallsBackToRasterization(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 50">` +
			`<rect width="100" height="50" fill="#ff0000"/></svg>`))
	}))
	defer srv.Close()

	base, err := nurl.Parse(srv.URL)
	require.NoError(t, err)

	blocks := []block{{kind: blockImage, imageURL: srv.URL + "/plot.svg"}}

	fetchImages(context.Background(), blocks, base)

	require.NotNil(t, blocks[0].img)
	assert.Equal(t, 100, blocks[0].img.Bounds().Dx())
	assert.Equal(t, 50, blocks[0].img.Bounds().Dy())

	r, g, b, a := blocks[0].img.At(50, 25).RGBA()
	assert.Equal(t, []uint32{0xffff, 0x0, 0x0, 0xffff}, []uint32{r, g, b, a}, "the rect fill is painted")
}

func TestDecodeSVG_BoundsOversizedViewBox(t *testing.T) {
	t.Parallel()

	img := decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 8192 4096"></svg>`))

	require.NotNil(t, img)
	assert.Equal(t, maxSVGRasterPx, img.Bounds().Dx())
	assert.Equal(t, maxSVGRasterPx/2, img.Bounds().Dy(), "aspect ratio is preserved")
}

func TestDecodeSVG_ToleratesImportantInStyles(t *testing.T) {
	t.Parallel()

	img := decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 50">` +
		`<rect width="100" height="50" style="fill:#ff0000 !important;stroke:#6f9bcb !important"/></svg>`))

	require.NotNil(t, img, "Mermaid-style !important declarations must not fail the parse")

	r, g, b, a := img.At(50, 25).RGBA()
	assert.Equal(t, []uint32{0xffff, 0x0, 0x0, 0xffff}, []uint32{r, g, b, a}, "the rect fill is painted")
}

func TestDecodeSVG_RejectsGarbage(t *testing.T) {
	t.Parallel()

	assert.Nil(t, decodeSVG([]byte("<html><body>404 not found</body></html>")))
	assert.Nil(t, decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)))
}

func TestBoundImage_DownscalesToFitBox(t *testing.T) {
	t.Parallel()

	bounded := boundImage(image.NewRGBA(image.Rect(0, 0, 2560, 1440)))

	assert.Equal(t, maxRetainedPx, bounded.Bounds().Dx())
	assert.Equal(t, 288, bounded.Bounds().Dy(), "aspect ratio is preserved")
}

func TestBoundImage_TallImageBoundsHeight(t *testing.T) {
	t.Parallel()

	bounded := boundImage(image.NewRGBA(image.Rect(0, 0, 400, 2000)))

	assert.Equal(t, 102, bounded.Bounds().Dx())
	assert.Equal(t, maxRetainedPx, bounded.Bounds().Dy())
}

func TestBoundImage_KeepsSmallImagesUntouched(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 100, 400))

	assert.Same(t, src, boundImage(src))
}
