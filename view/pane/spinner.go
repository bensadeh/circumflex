package pane

import (
	"image/color"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/bensadeh/circumflex/style"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const spinnerFrameDuration = 250 * time.Millisecond

// lastSpinnerColor tracks which color index was used last so the next spinner
// never repeats it. Mutex-guarded: spinners are minted at model construction
// as well as from the update loop, and parallel tests do both concurrently.
var (
	lastSpinnerColorMu sync.Mutex
	lastSpinnerColor   = -1
)

// UpdateSpinner advances the animation and reschedules the next tick only
// while active, so a stopped spinner's tick chain dies out.
func UpdateSpinner(sp spinner.Model, msg spinner.TickMsg, active bool) (spinner.Model, tea.Cmd) {
	next, cmd := sp.Update(msg)
	if !active {
		return next, nil
	}

	return next, cmd
}

func NewSpinner() spinner.Model {
	sp := spinner.New()
	sp.Spinner = starSpinner()
	sp.Style = lipgloss.NewStyle()

	return sp
}

func starSpinner() spinner.Spinner {
	colors := []color.Color{style.HeaderC(), style.HeaderL(), style.HeaderX()}
	s := lipgloss.NewStyle().Foreground(colors[nextSpinnerColor(len(colors))])

	// Every glyph must be East Asian Narrow: ambiguous-width ones (· U+00B7,
	// ✽ U+273D, ✳ U+2733) render double-width from a fallback font on some
	// terminals, so the glyph wobbled horizontally whenever the animation
	// crossed width classes.
	chars := []string{"∙", "✻", "❋", "✶", "✻", "✢", "✻", "✶", "❋", "✻"}
	frames := make([]string, len(chars))

	for i, ch := range chars {
		frames[i] = "   " + s.Render(ch) + "   "
	}

	return spinner.Spinner{
		Frames: frames,
		FPS:    spinnerFrameDuration,
	}
}

// nextSpinnerColor picks a random color index that differs from the previous
// pick, so consecutive spinners never share a color.
func nextSpinnerColor(count int) int {
	lastSpinnerColorMu.Lock()
	defer lastSpinnerColorMu.Unlock()

	var pick int
	if lastSpinnerColor == -1 {
		pick = rand.IntN(count)
	} else {
		// Pick from the indices that aren't lastSpinnerColor.
		offset := 1 + rand.IntN(count-1)
		pick = (lastSpinnerColor + offset) % count
	}

	lastSpinnerColor = pick

	return pick
}
