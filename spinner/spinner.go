package spinner

import (
	"clx/style"
	"image/color"
	"math/rand/v2"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
)

const frameDuration = 250 * time.Millisecond

// Random returns a randomly selected spinner animation.
func Random() spinner.Spinner {
	return star()
}

func star() spinner.Spinner {
	colors := []color.Color{style.HeaderC(), style.HeaderL(), style.HeaderX()}
	clr := colors[rand.IntN(len(colors))]
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
