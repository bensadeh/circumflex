package article

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	nurl "net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchImages_TinyImageIsDecorativeNotFailed(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w2, h := 100, 100

		switch r.URL.Path {
		case "/divider.png":
			w2, h = 10, 10
		case "/badge.png":
			w2, h = 129, 28
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w2, h))); err != nil {
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
		{kind: blockImage, imageURL: srv.URL + "/badge.png"},
	}

	fetchImages(context.Background(), blocks, base)

	assert.Zero(t, blocks[0].imgSize)
	assert.True(t, blocks[0].decorative, "below the size floors marks the block decorative")

	assert.NotZero(t, blocks[1].imgSize)
	assert.False(t, blocks[1].decorative)

	assert.Zero(t, blocks[2].imgSize)
	assert.True(t, blocks[2].decorative, "a repo badge is a short strip, not an image")
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

	require.NotZero(t, blocks[0].imgSize, "a terminal with graphics support renders the figure's pixels")
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

	assert.NotZero(t, blocks[0].imgSize, "same-origin image requests carry the page URL as Referer")
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
		{
			kind:    blockImage,
			imgSize: image.Pt(100, 100),
			kitty:   &kittyImage{png: []byte("png-bytes"), id: 7},
		},
	}}
	assert.True(t, decoded.HasImages(true))
	assert.False(t, decoded.HasImages(false), "without graphics support there is nothing to toggle")

	undecoded := &Parsed{blocks: []block{{kind: blockImage, imageURL: "https://example.com/a.png"}}}
	assert.False(t, undecoded.HasImages(true), "a failed fetch leaves nothing to toggle")

	assert.False(t, NewParsedFromHTML("<p>text only</p>").HasImages(true))
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

	require.NotNil(t, blocks[0].kitty)
	assert.Equal(t, image.Pt(maxImagePx, maxImagePx/2), blocks[0].imgSize, "the raster is drawn at the ceiling")
	assert.Equal(t, 100, blocks[0].dispWidth, "sizing comes from the viewBox, not the raster")

	rasterized, err := png.Decode(bytes.NewReader(blocks[0].kitty.png))
	require.NoError(t, err)

	r, g, b, a := rasterized.At(1024, 512).RGBA()
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

	require.NotZero(t, blocks[0].imgSize, "a logo with a small-unit viewBox is not a tracking pixel")
	assert.False(t, blocks[0].decorative)
	assert.Equal(t, 210, blocks[0].dispWidth)

	assert.Zero(t, blocks[1].imgSize)
	assert.True(t, blocks[1].decorative, "without a declared width the viewBox is the only size signal")
}

func TestFetchImages_UndrawableSVGJudgedByDeclaredGeometry(t *testing.T) {
	t.Parallel()

	// Malformed path data panics canvas's parser, so both fixtures fail to
	// rasterize and only the declared geometry is left to judge them by.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/badge.svg" {
			_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="129" height="28">` +
				`<path d="M zz garbage"/></svg>`))

			return
		}

		_, _ = w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 829">` +
			`<path d="M zz garbage"/></svg>`))
	}))
	defer srv.Close()

	base, err := nurl.Parse(srv.URL)
	require.NoError(t, err)

	blocks := []block{
		{kind: blockImage, imageURL: srv.URL + "/badge.svg"},
		{kind: blockImage, imageURL: srv.URL + "/sponsors.svg"},
	}

	fetchImages(context.Background(), blocks, base)

	assert.Zero(t, blocks[0].imgSize)
	assert.True(t, blocks[0].decorative, "a badge stays a badge when rasterization fails")

	assert.Zero(t, blocks[1].imgSize)
	assert.False(t, blocks[1].decorative, "a large undrawable vector keeps its honest label")
}

func TestSVGDeclaredBox(t *testing.T) {
	t.Parallel()

	assert.Equal(t, image.Pt(800, 829),
		svgDeclaredBox([]byte(`<?xml version="1.0"?><svg viewBox="0 0 800 829" width="640">x</svg>`)),
		"viewBox wins over width/height when both are present")
	assert.Equal(t, image.Pt(129, 28),
		svgDeclaredBox([]byte(`<svg width="128.5px" height="28">x</svg>`)),
		"width/height fill in when there is no viewBox")
	assert.Equal(t, image.Point{},
		svgDeclaredBox([]byte(`<svg width="100%" height="28">x</svg>`)),
		"non-pixel lengths carry no geometry")
	assert.Equal(t, image.Point{}, svgDeclaredBox([]byte("\xff\xd8not xml")))
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
	assert.Equal(t, maxImagePx, img.Bounds().Dx(), "oversized viewBoxes rasterize down to the ceiling")
	assert.Equal(t, maxImagePx/2, img.Bounds().Dy(), "aspect ratio is preserved")
	assert.Equal(t, image.Pt(8192, 4096), viewBox)

	img, viewBox = decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 126.5 20.22"></svg>`))

	require.NotNil(t, img)
	assert.Equal(t, maxImagePx, img.Bounds().Dx(), "small viewBoxes rasterize up to the ceiling")
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

// A missing font panics canvas mid-parse; the shapes must still render rather
// than the whole figure collapsing to its label (the pre-canvas oksvg output).
func TestDecodeSVG_MissingFontRendersShapes(t *testing.T) {
	t.Parallel()

	img, viewBox := decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 50">` +
		`<rect width="100" height="50" fill="#00ff00"/>` +
		`<text x="50" y="25" font-family="NoSuchFontFamilyXYZ">label</text></svg>`))

	require.NotNil(t, img, "an unloadable font must not lose the shapes")
	assert.Equal(t, image.Pt(100, 50), viewBox)

	r, g, b, _ := img.At(img.Bounds().Dx()/2, img.Bounds().Dy()/2).RGBA()
	assert.Equal(t, []uint32{0x0, 0xffff, 0x0}, []uint32{r, g, b}, "the rect fill survives the text-strip retry")
}

// canvas flags an unknown unit (em/ex/rem) as an error but keeps a drawable
// canvas; decodeSVG must render it rather than treat the error as total failure.
func TestDecodeSVG_NonFatalErrorKeepsShapes(t *testing.T) {
	t.Parallel()

	img, _ := decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 50">` +
		`<rect width="100" height="50" fill="#ff0000"/>` +
		`<text x="50" y="25" font-size="1.2em">hi</text></svg>`))

	require.NotNil(t, img, "a non-fatal parse error must not drop the drawable shapes")

	r, g, b, _ := img.At(img.Bounds().Dx()/2, img.Bounds().Dy()/2).RGBA()
	assert.Equal(t, []uint32{0xffff, 0x0, 0x0}, []uint32{r, g, b})
}

// A vector with no declared box has no on-page size, so it stays a label even
// though canvas could rasterize it — sizing off the raster would blow it up.
func TestDecodeSVG_NoDeclaredGeometryDropped(t *testing.T) {
	t.Parallel()

	img, viewBox := decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg">` +
		`<rect width="10" height="10" fill="#f00"/></svg>`))
	assert.Nil(t, img, "no viewBox and no declared size means no on-page geometry")
	assert.Equal(t, image.Point{}, viewBox)

	img, _ = decodeSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" width="30mm" height="10mm">` +
		`<rect width="30" height="10" fill="#f00"/></svg>`))
	assert.Nil(t, img, "non-pixel physical units carry no on-page geometry")
}

func TestStripSVGText(t *testing.T) {
	t.Parallel()

	assert.Nil(t, stripSVGText([]byte(`<svg><rect/></svg>`)), "no text to strip")

	stripped := stripSVGText([]byte(`<svg><rect x="1"/><text x="2">a</text><text y="3">b</text></svg>`))
	assert.Equal(t, `<svg><rect x="1"/></svg>`, string(stripped), "every text element goes, shapes stay")
}

// The mutex in drawSVG keeps concurrent text-SVG rasterization off canvas's
// unsynchronized global font state; this exercises that path for `go test -race`.
func TestDecodeSVG_ConcurrentTextSVGs(t *testing.T) {
	t.Parallel()

	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 50">` +
		`<rect width="100" height="50" fill="#123456"/><text x="50" y="25">x</text></svg>`)

	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			img, _ := decodeSVG(svg)
			assert.NotNil(t, img)
		})
	}

	wg.Wait()
}

func TestFetchImages_RetainsHighResKittyCopy(t *testing.T) {
	t.Parallel()

	// Encode both payloads once, up front, rather than per request: this fetch
	// rides fetchImages' 8s wall-clock timeout, and re-encoding a 4096px source
	// inside the handler starves that budget when the parallel canvas SVG
	// rasterization tests are pinning the runner's cores.
	encode := func(size int) []byte {
		var buf bytes.Buffer
		require.NoError(t, png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, size, size/2))))

		return buf.Bytes()
	}

	served := encode(100)
	huge := encode(maxImagePx * 2)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/huge.png" {
			_, _ = w.Write(huge)

			return
		}

		_, _ = w.Write(served)
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
	assert.Equal(t, maxImagePx, reencoded.Bounds().Dx(), "oversized sources re-encode within the bound")
	assert.Equal(t, maxImagePx/2, reencoded.Bounds().Dy(), "aspect ratio is preserved")

	assert.Equal(t, image.Pt(maxImagePx*2, maxImagePx), blocks[1].imgSize,
		"the retained size is the source's, which the aspect ratio derives from")
}
