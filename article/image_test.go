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
