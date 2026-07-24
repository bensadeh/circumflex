package pane

import (
	"image/color"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/bensadeh/circumflex/style"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	spinnerTickDuration = 80 * time.Millisecond
	spinnerCyclePeriod  = 2 * time.Second
)

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
	//
	// Glyphs ordered by visual weight: the animation climbs the ladder and
	// descends again once per cycle.
	glyphs := []string{"∙", "✢", "✶", "✻", "❋"}

	// The bubbles spinner shows every frame for the same interval, so the
	// breathing rhythm is pre-sampled: each tick maps its position in the
	// cycle through a raised cosine to a rung on the ladder. The wave
	// flattens at its extremes, so the dot and the full bloom hold for
	// ~0.4s each while the transitional shapes flick past.
	ticksPerCycle := int(spinnerCyclePeriod / spinnerTickDuration)
	frames := make([]string, ticksPerCycle)

	for tick := range ticksPerCycle {
		phase := 2 * math.Pi * float64(tick) / float64(ticksPerCycle)
		rung := (1 - math.Cos(phase)) / 2 * float64(len(glyphs)-1)
		frames[tick] = "   " + s.Render(glyphs[int(math.Round(rung))]) + "   "
	}

	return spinner.Spinner{
		Frames: frames,
		FPS:    spinnerTickDuration,
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
