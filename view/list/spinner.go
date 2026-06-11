package list

import (
	"image/color"
	"math/rand/v2"
	"time"

	"github.com/bensadeh/circumflex/style"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
)

const spinnerFrameDuration = 250 * time.Millisecond

// lastSpinnerColor tracks which color index was used last so the next spinner
// never repeats it. Only touched from the Bubble Tea update goroutine.
var lastSpinnerColor = -1

func newSpinner() spinner.Model {
	sp := spinner.New()
	sp.Spinner = starSpinner()
	sp.Style = defaultStyles().Spinner

	return sp
}

func starSpinner() spinner.Spinner {
	colors := []color.Color{style.HeaderC(), style.HeaderL(), style.HeaderX()}

	var pick int
	if lastSpinnerColor == -1 {
		pick = rand.IntN(len(colors))
	} else {
		// Pick from the two indices that aren't lastSpinnerColor.
		offset := 1 + rand.IntN(len(colors)-1) // 1 or 2
		pick = (lastSpinnerColor + offset) % len(colors)
	}

	lastSpinnerColor = pick
	s := lipgloss.NewStyle().Foreground(colors[pick])

	chars := []string{"·", "✻", "✽", "✶", "✳", "✢", "✳", "✶", "✽", "✻"}
	frames := make([]string, len(chars))

	for i, ch := range chars {
		frames[i] = "   " + s.Render(ch) + "   "
	}

	return spinner.Spinner{
		Frames: frames,
		FPS:    spinnerFrameDuration,
	}
}
