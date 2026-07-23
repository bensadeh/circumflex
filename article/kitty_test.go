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
	// columns.
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
	return &block{
		kind:      blockImage,
		imgSize:   image.Pt(800, 400),
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

func TestCachedImagePartTracksCellSize(t *testing.T) {
	t.Parallel()

	b := kittyTestBlock()

	wide := cachedImagePart(b, 44, ImageOptions{Show: true, Kitty: true})
	assert.Contains(t, wide, string(kitty.Placeholder))

	square := cachedImagePart(b, 44, ImageOptions{Show: true, Kitty: true, CellWidth: 10, CellHeight: 10})
	assert.NotEqual(t, wide, square, "a font-size change re-renders rather than serving the stale grid")

	assert.Equal(t, wide, cachedImagePart(b, 44, ImageOptions{Show: true, Kitty: true}),
		"the original geometry re-renders too")
}

func TestImageRendersAsLabelWithoutGraphicsSupport(t *testing.T) {
	t.Parallel()

	b := kittyTestBlock()
	b.spans = []span{{text: "a caption"}}

	shown := renderImage(b, 44, ImageOptions{Show: true, Kitty: true})
	assert.Contains(t, shown, string(kitty.Placeholder), "the terminal composites the pixels")

	unsupported := xansi.Strip(renderImage(b, 44, ImageOptions{Show: true}))
	assert.Contains(t, unsupported, "●●● Image a caption",
		"a terminal that cannot draw the image describes it instead")
	assert.NotContains(t, unsupported, string(kitty.Placeholder))

	b.figure = true
	figure := xansi.Strip(renderImage(b, 44, ImageOptions{Show: true}))
	assert.Contains(t, figure, "▂▄▆ Figure a caption", "a known chart says so")

	hidden := xansi.Strip(renderImage(b, 44, ImageOptions{Kitty: true}))
	assert.Contains(t, hidden, "▂▄▆ Figure a caption")
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
