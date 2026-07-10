package pane

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// repaintDelay must outlast the flush of the redraw the resize itself
// triggers — one frame at Bubble Tea's 60fps cadence, with headroom.
const repaintDelay = 100 * time.Millisecond

// RepaintAfterGrow works around a renderer bug that surfaces when the
// terminal grows wider: ultraviolet resizes its model of the screen but
// backfills only newly added rows, never the columns a grow adds to existing
// rows (TerminalRenderer.Render's sync loop), so it records those columns as
// blank while the resize redraw put content there. Anything in them that
// later becomes blank diffs as unchanged and lingers on screen as ghost
// text. A full repaint after that redraw has flushed — once the model has
// its correct dimensions — resyncs the model with the screen.
func RepaintAfterGrow() tea.Cmd {
	return tea.Tick(repaintDelay, func(time.Time) tea.Msg { return tea.ClearScreen() })
}
