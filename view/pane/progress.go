package pane

import (
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
)

// Terminal progress bar via OSC 9;4 (supported by Ghostty, ConEmu and others;
// silently ignored by terminals that don't recognise the sequence).
//
// While a program runs, sequences ride progressCh into its message loop and
// leave through its own output, serialized with frame flushes. Writing to the
// terminal directly would race them: Bubble Tea flushes frames from its own
// goroutine, backpressure splits a frame across several writes, and a
// sequence landing between two chunks corrupts the terminal's parse — the
// cell-diff renderer then leaves ghost text it believes was repainted.
// ProgressOut serves tests and the final clear after the program exits.

var ProgressOut io.Writer = os.Stderr

// progressCh is wired by WireProgress; nil whenever no program is running.
var progressCh chan<- string

const progressClearSeq = "\033]9;4;0\a"

func SetProgressIndeterminate()  { emitProgress("\033]9;4;3;0\a") }
func SetProgressPercent(pct int) { emitProgress(fmt.Sprintf("\033]9;4;1;%d\a", pct)) }
func setProgressError()          { emitProgress("\033]9;4;2;100\a") }

func ClearProgress() { emitProgress(progressClearSeq) }

func emitProgress(seq string) {
	if progressCh != nil {
		// Progress is cosmetic: if the program stopped consuming, drop the
		// update rather than block.
		select {
		case progressCh <- seq:
		default:
		}

		return
	}

	_, _ = fmt.Fprint(ProgressOut, seq)
}

// SyncProgress settles the indicator for a finished fetch: an error stays
// visible for the status message lifetime, success clears it. Call only from
// the Update loop after the finish guard, so a stale fetch can never write
// over its successor's indicator.
func SyncProgress(err error) {
	if err != nil {
		setProgressError()

		return
	}

	ClearProgress()
}

// WireProgress routes progress sequences into p's message loop for the
// program's lifetime. The forwarder goroutine exists because Send would
// deadlock called from Update itself; once the program stops, Send is a
// no-op. The returned settle stops the forwarder and clears the indicator —
// call it after Run returns, when a direct write can no longer interleave
// with a frame; clearing there rather than in the quit paths guarantees the
// indicator never outlives the program, whatever the exit.
func WireProgress(p *tea.Program) (settle func()) {
	ch := make(chan string, 64)
	done := make(chan struct{})
	progressCh = ch

	go func() {
		for {
			select {
			case seq := <-ch:
				p.Send(tea.RawMsg{Msg: seq})
			case <-done:
				return
			}
		}
	}()

	return func() {
		close(done)

		progressCh = nil

		_, _ = fmt.Fprint(ProgressOut, progressClearSeq)
	}
}
