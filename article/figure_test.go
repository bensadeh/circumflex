package article

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnownFigure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want bool
	}{
		{name: "genre-led caption", text: "Line chart of training loss", want: true},
		{name: "genre with article and qualifier", text: "A Sankey diagram of energy flows", want: true},
		{name: "qualified plot", text: "Scatter plot of accuracy against size", want: true},
		{name: "histogram", text: "Histogram of scores", want: true},
		{name: "numbered figure", text: "Figure 3: Attention entropy by layer", want: true},
		{name: "abbreviated numbered figure", text: "Fig. 2a shows the raw results", want: true},
		{name: "photo", text: "A horse in a field", want: false},
		{name: "undeclared graphic", text: "Training loss over time", want: false},
		{name: "genre mentioned mid-sentence", text: "Man holding a chart", want: false},
		{name: "bare plot is land, not data", text: "Plot of land for sale", want: false},
		{name: "graphic is not graph", text: "Graphic novel cover art", want: false},
		{name: "figure-adjacent word", text: "Fighter jet at an air show", want: false},
		{name: "empty", text: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, knownFigure(tt.text))
		})
	}
}
