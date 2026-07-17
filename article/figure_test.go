package article

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnownFigure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		text string
		want bool
	}{
		{name: "described svg", src: "chart.svg", text: "Training loss", want: true},
		{name: "svg with query string", src: "chart.svg?v=2", text: "Training loss", want: true},
		{name: "undescribed svg stays image", src: "chart.svg", text: "", want: false},
		{name: "genre-led caption", src: "loss.png", text: "Line chart of training loss", want: true},
		{name: "genre with article and qualifier", src: "x.png", text: "A Sankey diagram of energy flows", want: true},
		{name: "qualified plot", src: "x.png", text: "Scatter plot of accuracy against size", want: true},
		{name: "histogram", src: "x.png", text: "Histogram of scores", want: true},
		{name: "numbered figure", src: "fig3.png", text: "Figure 3: Attention entropy by layer", want: true},
		{name: "abbreviated numbered figure", src: "x.png", text: "Fig. 2a shows the raw results", want: true},
		{name: "photo", src: "horse.jpg", text: "A horse in a field", want: false},
		{name: "genre mentioned mid-sentence", src: "x.jpg", text: "Man holding a chart", want: false},
		{name: "bare plot is land, not data", src: "x.jpg", text: "Plot of land for sale", want: false},
		{name: "graphic is not graph", src: "x.jpg", text: "Graphic novel cover art", want: false},
		{name: "figure-adjacent word", src: "x.jpg", text: "Fighter jet at an air show", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, knownFigure(tt.src, tt.text))
		})
	}
}
