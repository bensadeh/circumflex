package pane

import (
	"strings"
	"sync"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/graphics"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

// DetectGraphics probes the terminal for the Kitty graphics protocol and its
// cell pixel size, enabling high-resolution reader images. Terminals that
// answer confirm through HandleGraphicsReport; no answer leaves the
// half-block art. Terminals the probe would misfire on return nil.
func DetectGraphics() tea.Cmd {
	if !graphics.ShouldQuery() {
		return nil
	}

	return tea.Raw(graphics.QuerySeq())
}

// QueryCellSize re-asks for the cell pixel size on a resize, so a font-size
// change re-derives image geometry. Nothing to keep honest before the
// graphics probe succeeded.
func QueryCellSize() tea.Cmd {
	if !graphics.Enabled() {
		return nil
	}

	return tea.Raw(graphics.CellSizeQuerySeq())
}

// HandleGraphicsReport records a terminal's answer to the graphics probes
// and reports whether the state changed — the caller then repaints an open
// reader so its images pick up the new mode or geometry.
func HandleGraphicsReport(msg tea.Msg) bool {
	switch msg := msg.(type) {
	case uv.KittyGraphicsEvent:
		return graphics.IsQueryReply(msg.Options.ID) && graphics.Enable()

	case uv.CellSizeEvent:
		return graphics.SetCellSize(msg.Width, msg.Height)
	}

	return false
}

// Image work reaches the terminal through a single worker, in submission
// order. Order is a correctness constraint: a placement-only batch builds in
// microseconds while a transmission base64-encodes megabytes of PNG, so
// delivered as independent commands they arrive in either order — and a
// transmission landing after the placement that corrects it pins geometry
// the delta protocol already considers settled, leaving the image sized
// against the wrong grid until the page is left. The worker sends through
// the program's message loop, serialized with frame flushes like the
// progress OSC.
var (
	kittyMu    sync.Mutex
	kittyQueue [][]article.KittyWork

	// kittyKick is wired by WireGraphics before its program starts and never
	// written again; the pre-start write is what makes the unsynchronized
	// reads safe. Nil when no program was wired — tests, which drive
	// wireGraphics directly.
	kittyKick chan<- struct{}
)

// EmitKittyWork queues a render's image work — PNGs not yet transmitted,
// placements whose geometry changed — for ordered delivery to the terminal.
// Call from the update loop only: submission order must match the render
// order the delta protocol's bookkeeping saw.
func EmitKittyWork(work []article.KittyWork) {
	if len(work) == 0 || kittyKick == nil {
		return
	}

	kittyMu.Lock()

	kittyQueue = append(kittyQueue, work)

	kittyMu.Unlock()

	// Non-blocking: the update loop must never wait on the worker, and one
	// pending kick is enough — the worker drains the whole queue per kick.
	select {
	case kittyKick <- struct{}{}:
	default:
	}
}

// WireGraphics routes image work into p's message loop. Call it before Run;
// the returned stop ends the worker once the program has exited.
func WireGraphics(p *tea.Program) (stop func()) {
	return wireGraphics(func(seq string) {
		p.Send(tea.RawMsg{Msg: seq})
	})
}

func wireGraphics(send func(string)) (stop func()) {
	kick := make(chan struct{}, 1)
	done := make(chan struct{})

	kittyMu.Lock()
	kittyQueue = nil
	kittyMu.Unlock()

	kittyKick = kick

	go func() {
		for {
			select {
			case <-kick:
			case <-done:
				return
			}

			for {
				kittyMu.Lock()
				if len(kittyQueue) == 0 {
					kittyMu.Unlock()

					break
				}

				work := kittyQueue[0]
				kittyQueue = kittyQueue[1:]
				kittyMu.Unlock()

				send(kittySeq(work))
			}
		}
	}()

	return func() { close(done) }
}

func kittySeq(work []article.KittyWork) string {
	var sb strings.Builder

	for _, w := range work {
		if w.PNG != nil {
			sb.WriteString(graphics.TransmitSeq(w.ID, w.PNG, w.Cols, w.Rows))
		} else {
			sb.WriteString(graphics.PlacementSeq(w.ID, w.Cols, w.Rows))
		}
	}

	return sb.String()
}
