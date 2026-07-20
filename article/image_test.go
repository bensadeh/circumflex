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

func TestFetchImages_FetchesKnownFigures(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		var buf bytes.Buffer
		if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 100, 100))); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		_, _ = w.Write(buf.Bytes())
	}))
	defer srv.Close()

	base, err := nurl.Parse(srv.URL)
	require.NoError(t, err)

	blocks := []block{
		{kind: blockImage, imageURL: srv.URL + "/chart.png", figure: true},
	}

	fetchImages(context.Background(), blocks, base)

	require.NotNil(t, blocks[0].img, "a Kitty-tier terminal renders the figure's pixels")
	assert.NotNil(t, blocks[0].kitty)
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
	assert.True(t, decoded.HasImages(false))

	undecoded := &Parsed{blocks: []block{{kind: blockImage, imageURL: "https://example.com/a.png"}}}
	assert.False(t, undecoded.HasImages(false), "a failed fetch leaves nothing to toggle")

	figure := &Parsed{blocks: []block{{
		kind:   blockImage,
		figure: true,
		img:    image.NewRGBA(image.Rect(0, 0, 100, 100)),
		kitty:  &kittyImage{png: []byte("png-bytes"), id: 7},
	}}}
	assert.True(t, figure.HasImages(true))
	assert.False(t, figure.HasImages(false), "below the Kitty tier the toggle never reveals a figure")

	assert.False(t, NewParsedFromHTML("<p>text only</p>").HasImages(false))
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
	assert.Equal(t, tierMaxPx[tierHalfBlock], blocks[0].img.Bounds().Dx(), "the retained copy keeps its own bound")
	assert.Equal(t, tierMaxPx[tierHalfBlock]/2, blocks[0].img.Bounds().Dy())
	assert.Equal(t, 100, blocks[0].dispWidth, "sizing comes from the viewBox, not the raster")

	r, g, b, a := blocks[0].img.At(256, 128).RGBA()
	assert.Equal(t, []uint32{0xffff, 0x0, 0x0, 0xffff}, []uint32{r, g, b, a}, "the rect fill is painted")
}

func TestFetchImages_SmallUnitViewBoxLogoJudgedByDeclaredWidth(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 126.5 20.22">` +
			`<rect width="126.5" height="20.22" fill="#5865f2"/></svg>`))
	}))
	defer srv.Close()

	base, err := nurl.Parse(srv.URL)
	require.NoError(t, err)

	blocks := []block{
		{kind: blockImage, imageURL: srv.URL + "/discord.svg", dispWidth: 210},
		{kind: blockImage, imageURL: srv.URL + "/discord.svg"},
	}

	fetchImages(context.Background(), blocks, base)

	require.NotNil(t, blocks[0].img, "a logo with a small-unit viewBox is not a tracking pixel")
	assert.False(t, blocks[0].decorative)
	assert.Equal(t, 210, blocks[0].dispWidth)

	assert.Nil(t, blocks[1].img)
	assert.True(t, blocks[1].decorative, "without a declared width the viewBox is the only size signal")
}

func TestSVGDisplaySize(t *testing.T) {
	t.Parallel()

	assert.Equal(t, image.Pt(210, 33), svgDisplaySize(image.Pt(127, 20), 210),
		"declared width scales through the viewBox aspect")
	assert.Equal(t, image.Pt(127, 20), svgDisplaySize(image.Pt(127, 20), 0),
		"no declared width falls back to the viewBox")
}

func TestDecodeSVG_RasterizesAtKittyCeiling(t *testing.T) {
	t.Parallel()

	img, viewBox := decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 8192 4096"></svg>`))

	require.NotNil(t, img)
	assert.Equal(t, tierMaxPx[tierKitty], img.Bounds().Dx(), "oversized viewBoxes rasterize down to the ceiling")
	assert.Equal(t, tierMaxPx[tierKitty]/2, img.Bounds().Dy(), "aspect ratio is preserved")
	assert.Equal(t, image.Pt(8192, 4096), viewBox)

	img, viewBox = decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 126.5 20.22"></svg>`))

	require.NotNil(t, img)
	assert.Equal(t, tierMaxPx[tierKitty], img.Bounds().Dx(), "small viewBoxes rasterize up to the ceiling")
	assert.Equal(t, 327, img.Bounds().Dy())
	assert.Equal(t, image.Pt(127, 20), viewBox, "the viewBox rounds to the nearest unit")
}

func TestDecodeSVG_ToleratesImportantInStyles(t *testing.T) {
	t.Parallel()

	img, _ := decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 50">` +
		`<rect width="100" height="50" style="fill:#ff0000 !important;stroke:#6f9bcb !important"/></svg>`))

	require.NotNil(t, img, "Mermaid-style !important declarations must not fail the parse")

	r, g, b, a := img.At(512, 256).RGBA()
	assert.Equal(t, []uint32{0xffff, 0x0, 0x0, 0xffff}, []uint32{r, g, b, a}, "the rect fill is painted")
}

func TestDecodeSVG_RejectsGarbage(t *testing.T) {
	t.Parallel()

	img, _ := decodeSVG([]byte("<html><body>404 not found</body></html>"))
	assert.Nil(t, img)

	img, _ = decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`))
	assert.Nil(t, img)
}

func TestFetchImages_RetainsHighResKittyCopy(t *testing.T) {
	t.Parallel()

	var served []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size := 100
		if r.URL.Path == "/huge.png" {
			size = tierMaxPx[tierKitty] * 2
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, size, size/2))); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if r.URL.Path != "/huge.png" {
			served = buf.Bytes()
		}

		_, _ = w.Write(buf.Bytes())
	}))
	defer srv.Close()

	base, err := nurl.Parse(srv.URL)
	require.NoError(t, err)

	blocks := []block{
		{kind: blockImage, imageURL: srv.URL + "/photo.png"},
		{kind: blockImage, imageURL: srv.URL + "/huge.png"},
	}

	fetchImages(context.Background(), blocks, base)

	require.NotNil(t, blocks[0].kitty)
	assert.Equal(t, served, blocks[0].kitty.png, "an in-bounds PNG passes through byte-for-byte")
	assert.Positive(t, blocks[0].kitty.id)

	require.NotNil(t, blocks[1].kitty)

	reencoded, err := png.Decode(bytes.NewReader(blocks[1].kitty.png))
	require.NoError(t, err)
	assert.Equal(t, tierMaxPx[tierKitty], reencoded.Bounds().Dx(), "oversized sources re-encode within the kitty bound")
	assert.Equal(t, tierMaxPx[tierKitty]/2, reencoded.Bounds().Dy(), "aspect ratio is preserved")

	assert.Equal(t, tierMaxPx[tierHalfBlock], blocks[1].img.Bounds().Dx(), "the half-block fallback keeps its own tighter bound")
}

func TestBoundImage_DownscalesToFitBox(t *testing.T) {
	t.Parallel()

	bounded := boundImage(image.NewRGBA(image.Rect(0, 0, 2560, 1440)))

	assert.Equal(t, tierMaxPx[tierHalfBlock], bounded.Bounds().Dx())
	assert.Equal(t, 288, bounded.Bounds().Dy(), "aspect ratio is preserved")
}

func TestBoundImage_TallImageBoundsHeight(t *testing.T) {
	t.Parallel()

	bounded := boundImage(image.NewRGBA(image.Rect(0, 0, 400, 2000)))

	assert.Equal(t, 102, bounded.Bounds().Dx())
	assert.Equal(t, tierMaxPx[tierHalfBlock], bounded.Bounds().Dy())
}

func TestBoundImage_KeepsSmallImagesUntouched(t *testing.T) {
	t.Parallel()

	src := image.NewRGBA(image.Rect(0, 0, 100, 400))

	assert.Same(t, src, boundImage(src))
}
