package spinner

import (
	"image/color"
	"math/rand/v2"
	"time"

	"github.com/bensadeh/circumflex/style"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
)

const frameDuration = 250 * time.Millisecond

// lastColor tracks which color index was used last so the next call never
// repeats. A value of -1 means no color has been chosen yet.
var lastColor = -1

func Random() spinner.Spinner {
	return star()
}

func star() spinner.Spinner {
	colors := []color.Color{style.HeaderC(), style.HeaderL(), style.HeaderX()}

	var pick int
	if lastColor == -1 {
		pick = rand.IntN(len(colors))
	} else {
		// Pick from the two indices that aren't lastColor.
		offset := 1 + rand.IntN(len(colors)-1) // 1 or 2
		pick = (lastColor + offset) % len(colors)
	}

	lastColor = pick
	clr := colors[pick]
	s := lipgloss.NewStyle().Foreground(clr)

	chars := []string{"·", "✻", "✽", "✶", "✳", "✢", "✳", "✶", "✽", "✻"}
	frames := make([]string, len(chars))

	for i, ch := range chars {
		frames[i] = "   " + s.Render(ch) + "   "
	}

	return spinner.Spinner{
		Frames: frames,
		FPS:    frameDuration,
	}
}
