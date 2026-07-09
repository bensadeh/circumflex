package article

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
