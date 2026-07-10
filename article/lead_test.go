package article

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const featuredImagePage = `<html><body>
	<header><img src="https://example.com/uploads/hero-976x603.jpg" width="976" alt="the hero" class="entry__image wp-post-image"></header>
	<div><p>Body text.</p></div>
</body></html>`

func TestRestoreLeadImage_PrependsDroppedFeaturedImage(t *testing.T) {
	t.Parallel()

	blocks := []block{{kind: blockParagraph, spans: []span{{text: "Body text."}}}}

	restored := restoreLeadImage([]byte(featuredImagePage), blocks)

	require.Len(t, restored, 2)
	assert.Equal(t, blockImage, restored[0].kind)
	assert.Equal(t, "https://example.com/uploads/hero-976x603.jpg", restored[0].imageURL)
	assert.Equal(t, "the hero", restored[0].plainText())
	assert.Equal(t, 976, restored[0].dispWidth)
}

func TestRestoreLeadImage_SkipsWhenSizeVariantAlreadyPresent(t *testing.T) {
	t.Parallel()

	blocks := []block{
		{kind: blockImage, imageURL: "https://example.com/uploads/hero-scaled.jpg"},
		{kind: blockParagraph, spans: []span{{text: "Body text."}}},
	}

	restored := restoreLeadImage([]byte(featuredImagePage), blocks)

	require.Len(t, restored, 2, "size variants of the same upload must not duplicate")
	assert.Equal(t, "https://example.com/uploads/hero-scaled.jpg", restored[0].imageURL)
}

func TestRestoreLeadImage_NoFeaturedImageLeavesBlocksAlone(t *testing.T) {
	t.Parallel()

	blocks := []block{{kind: blockParagraph, spans: []span{{text: "Body text."}}}}

	restored := restoreLeadImage([]byte(`<html><body><img src="a.png"><p>Body text.</p></body></html>`), blocks)

	assert.Len(t, restored, 1)
}
