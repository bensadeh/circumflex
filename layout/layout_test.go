package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrame_NarrowIsSinglePane(t *testing.T) {
	t.Parallel()

	f := Frame{Width: 100, Height: 40, Wide: false}

	assert.Equal(t, 100, f.ListWidth())
	assert.Equal(t, 100, f.DetailWidth())
}

func TestFrame_WideSplitsWithDivider(t *testing.T) {
	t.Parallel()

	f := Frame{Width: 181, Height: 40, Wide: true}

	assert.Equal(t, (181-PaneDividerWidth)/2, f.ListWidth())
	assert.Equal(t, 181, f.ListWidth()+PaneDividerWidth+f.DetailWidth(),
		"panes plus divider fill the whole width exactly")
}

func TestFrame_PaneContentHeightReservesChrome(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 40-PaneChromeHeight, Frame{Height: 40}.PaneContentHeight())
	assert.Equal(t, 0, Frame{Height: 2}.PaneContentHeight(), "never negative")
}

func TestReaderContentWidthCapsAtMax(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 80, ReaderContentWidth(400, 80), "capped by max article width")
	assert.Equal(t, 50-2*ReaderViewLeftMargin, ReaderContentWidth(50, 80))
}
