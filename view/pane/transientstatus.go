package pane

import (
	"time"

	"github.com/bensadeh/circumflex/view/message"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	StatusMessageShort = 2 * time.Second
	StatusMessageLong  = 3 * time.Second
)

// TransientStatus is a footer status message that expires after a lifetime;
// the generation guards a stale timer from clearing a newer message.
type TransientStatus struct {
	message    string
	generation int
}

func (t *TransientStatus) Message() string { return t.message }

// Set shows msg and returns the command that expires it after d.
func (t *TransientStatus) Set(msg string, d time.Duration) tea.Cmd {
	t.message = msg
	t.generation++

	gen := t.generation

	return tea.Tick(d, func(time.Time) tea.Msg {
		return message.StatusMessageTimeout{Generation: gen}
	})
}

// SetPermanent shows msg with no expiry: it stays until the next Set,
// SetPermanent or Clear. The generation bump invalidates pending expiry
// timers — the message may be modal text like the favorites prompt, which
// an earlier message's timer must not blank.
func (t *TransientStatus) SetPermanent(msg string) {
	t.message = msg
	t.generation++
}

// Clear hides the message immediately and invalidates pending expiries.
func (t *TransientStatus) Clear() {
	t.message = ""
	t.generation++
}

// Expire clears the message if gen is current — i.e. the timer that fired
// belongs to the visible message — and reports whether it did.
func (t *TransientStatus) Expire(gen int) bool {
	if gen != t.generation {
		return false
	}

	t.Clear()

	return true
}

// CancelledStatus is the message shown when an in-flight operation is
// cancelled by hand.
func CancelledStatus() string {
	return lipgloss.NewStyle().Faint(true).Render("Cancelled")
}
