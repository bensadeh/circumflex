package pane

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/graphics"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectGraphics(t *testing.T) {
	clearTermEnv := func(t *testing.T) {
		t.Helper()
		t.Setenv("TERM", "")
		t.Setenv("TERM_PROGRAM", "")
		t.Setenv("TMUX", "")
	}

	t.Run("unknown terminals are probed", func(t *testing.T) {
		clearTermEnv(t)
		t.Setenv("TERM", "xterm-256color")

		assert.NotNil(t, DetectGraphics(), "no answer costs nothing; an answer unlocks high-res images")
	})

	t.Run("Apple Terminal is never probed", func(t *testing.T) {
		clearTermEnv(t)
		t.Setenv("TERM_PROGRAM", "Apple_Terminal")

		assert.Nil(t, DetectGraphics(), "an APC probe would print as garbage in Terminal.app")
	})

	t.Run("WezTerm is never probed", func(t *testing.T) {
		clearTermEnv(t)
		t.Setenv("TERM_PROGRAM", "WezTerm")

		assert.Nil(t, DetectGraphics(), "it answers the probe but draws no Unicode placeholders")
	})
}

// Enabling is package-global and sticky, so the disabled-state assertions
// run before the report flips it — deliberately serial, like the underline
// detection tests.
func TestHandleGraphicsReport(t *testing.T) {
	assert.False(t, HandleGraphicsReport(uv.KeyPressEvent{}), "unrelated events are not graphics reports")

	assert.False(t, HandleGraphicsReport(uv.KittyGraphicsEvent{Options: kitty.Options{ID: 7}}),
		"a stray graphics response is not the probe echo")
	assert.Nil(t, QueryCellSize(), "nothing to keep honest before the probe succeeded")

	assert.True(t, HandleGraphicsReport(uv.KittyGraphicsEvent{Options: kitty.Options{ID: 31}}))
	assert.True(t, graphics.Enabled())
	assert.False(t, HandleGraphicsReport(uv.KittyGraphicsEvent{Options: kitty.Options{ID: 31}}),
		"a repeated answer changes nothing")

	assert.True(t, HandleGraphicsReport(uv.CellSizeEvent{Width: 10, Height: 22}))
	assert.False(t, HandleGraphicsReport(uv.CellSizeEvent{Width: 10, Height: 22}),
		"an unchanged cell size changes nothing")

	assert.NotNil(t, QueryCellSize(), "resizes re-ask once a graphics terminal answered")
}

// The pipeline is package-global (kittyKick), so this runs serial like the
// other graphics-state tests.
func TestKittyWorkPipeline(t *testing.T) {
	t.Setenv("TMUX", "")

	var (
		mu  sync.Mutex
		got []string
	)

	stop := wireGraphics(func(seq string) {
		mu.Lock()
		defer mu.Unlock()

		got = append(got, seq)
	})
	defer stop()

	EmitKittyWork(nil)
	EmitKittyWork([]article.KittyWork{{ID: 5, PNG: []byte("png"), Cols: 4, Rows: 2}})
	EmitKittyWork([]article.KittyWork{{ID: 5, Cols: 6, Rows: 3}})

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()

		return len(got) == 2
	}, time.Second, time.Millisecond, "empty work is dropped, the rest delivered")

	mu.Lock()
	defer mu.Unlock()

	assert.Contains(t, got[0], "a=T", "the transmission leaves first")
	assert.Contains(t, got[0], "c=4,r=2")
	assert.True(t, strings.Contains(got[1], "a=p") && !strings.Contains(got[1], "a=T"),
		"a geometry-only batch replaces the placement, no pixels resent")
	assert.Contains(t, got[1], "c=6,r=3")
}
