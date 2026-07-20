package article

import (
	"image"
	"strings"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKittyGrid(t *testing.T) {
	t.Parallel()

	// A full-width square image on default 1:2 cells: half as many rows as
	// columns, like the half-block art.
	cols, rows := kittyGrid(640, 800, 800, 70, 0, 0)
	assert.Equal(t, 70, cols)
	assert.Equal(t, 35, rows)

	// Square cells keep a square image square.
	cols, rows = kittyGrid(640, 800, 800, 30, 10, 10)
	assert.Equal(t, 30, cols)
	assert.Equal(t, 30, rows)

	// A pathologically tall image caps at maxImageRows and narrows to keep
	// its aspect ratio, so it can still scroll past.
	cols, rows = kittyGrid(640, 100, 4000, 70, 0, 0)
	assert.Equal(t, maxImageRows, rows)
	assert.Equal(t, 2, cols)
}

func kittyTestBlock() *block {
	img := image.NewRGBA(image.Rect(0, 0, 800, 400))

	// Opaque pixels, so the half-block fallback paints cells rather than
	// transparent blanks.
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 0xff
	}

	return &block{
		kind:      blockImage,
		img:       img,
		kitty:     &kittyImage{png: []byte("png-bytes"), id: 42},
		dispWidth: 640,
	}
}

func TestRenderKittyArt(t *testing.T) {
	t.Parallel()

	b := kittyTestBlock()

	art := renderKittyArt(b, 40, 0, 0)
	require.NotEmpty(t, art)

	lines := strings.Split(art, "\n")
	assert.Len(t, lines, 10, "40 cols at 2:1 aspect on 1:2 cells")

	for _, line := range lines {
		assert.Equal(t, 40, xansi.StringWidth(line), "placeholder cells measure like ordinary text")
	}

	assert.True(t, strings.HasPrefix(lines[0], "\x1b[38;5;42m"), "the indexed foreground carries the image ID")
	assert.True(t, strings.HasPrefix(lines[1], "\x1b[38;5;42m"+
		string(kitty.Placeholder)+string(kitty.Diacritic(1))+string(kitty.Diacritic(0))),
		"each row opens with its row and column diacritics")

	assert.Equal(t, 40, b.kitty.wantCols, "the render records the geometry it laid down")
	assert.Equal(t, 10, b.kitty.wantRows)
}

func TestCachedImagePartSwitchesWithMode(t *testing.T) {
	t.Parallel()

	b := kittyTestBlock()

	kittyPart := cachedImagePart(b, 44, ImageOptions{Show: true, Kitty: true})
	assert.Contains(t, kittyPart, string(kitty.Placeholder))

	blockPart := cachedImagePart(b, 44, ImageOptions{Show: true})
	assert.NotContains(t, blockPart, string(kitty.Placeholder), "half-block art for terminals without graphics")
	assert.Contains(t, blockPart, "▀")

	assert.Equal(t, kittyPart, cachedImagePart(b, 44, ImageOptions{Show: true, Kitty: true}),
		"switching back re-renders rather than serving the stale mode")
}

func TestFigureArtOnlyAtKittyTier(t *testing.T) {
	t.Parallel()

	b := kittyTestBlock()
	b.figure = true
	b.spans = []span{{text: "a caption"}}

	kittyPart := renderImage(b, 44, ImageOptions{Show: true, Kitty: true})
	assert.Contains(t, kittyPart, string(kitty.Placeholder), "high resolution keeps the chart legible")

	halfBlockPart := xansi.Strip(renderImage(b, 44, ImageOptions{Show: true}))
	assert.Contains(t, halfBlockPart, "▂▄▆ Figure a caption", "half-block art would smear the chart")
	assert.NotContains(t, halfBlockPart, "▀")

	hiddenPart := xansi.Strip(renderImage(b, 44, ImageOptions{Kitty: true}))
	assert.Contains(t, hiddenPart, "▂▄▆ Figure a caption")
}

func TestPendingKittyWork(t *testing.T) {
	t.Parallel()

	b := kittyTestBlock()
	p := &Parsed{blocks: []block{*b}}

	assert.Empty(t, p.PendingKittyWork(), "nothing rendered placeholders yet, nothing owed")

	renderBlocks(p.blocks, 44, 44, ImageOptions{Show: true, Kitty: true})

	work := p.PendingKittyWork()
	require.Len(t, work, 1)
	assert.Equal(t, 42, work[0].ID)
	assert.Equal(t, []byte("png-bytes"), work[0].PNG, "first settle transmits the pixels")
	assert.Equal(t, 40, work[0].Cols)
	assert.Equal(t, 10, work[0].Rows)

	assert.Empty(t, p.PendingKittyWork(), "settled state owes nothing")

	// A narrower layout changes the grid; only the placement travels.
	renderBlocks(p.blocks, 24, 24, ImageOptions{Show: true, Kitty: true})

	work = p.PendingKittyWork()
	require.Len(t, work, 1)
	assert.Nil(t, work[0].PNG, "the terminal already holds the pixels")
	assert.Equal(t, 20, work[0].Cols)

	// A hidden render records geometry too, so the terminal's placement
	// tracks the layout while images are off and showing again owes nothing.
	renderBlocks(p.blocks, 44, 44, ImageOptions{Kitty: true})

	work = p.PendingKittyWork()
	require.Len(t, work, 1)
	assert.Nil(t, work[0].PNG)
	assert.Equal(t, 40, work[0].Cols)

	renderBlocks(p.blocks, 44, 44, ImageOptions{Show: true, Kitty: true})
	assert.Empty(t, p.PendingKittyWork(), "showing at the settled width owes nothing")
}

func TestHiddenRenderTransmitsAheadOfFirstShow(t *testing.T) {
	t.Parallel()

	b := kittyTestBlock()
	p := &Parsed{blocks: []block{*b}}

	// The page opens with images hidden: no placeholder cells, but the
	// pixels travel now, while nothing on screen is waiting for them.
	renderBlocks(p.blocks, 44, 44, ImageOptions{Kitty: true})

	work := p.PendingKittyWork()
	require.Len(t, work, 1)
	assert.Equal(t, []byte("png-bytes"), work[0].PNG)
	assert.Equal(t, 40, work[0].Cols)
	assert.Equal(t, 10, work[0].Rows)

	renderBlocks(p.blocks, 44, 44, ImageOptions{Show: true, Kitty: true})
	assert.Empty(t, p.PendingKittyWork(),
		"the first show composites against pixels the terminal already holds")
}
