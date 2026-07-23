package pane

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveWindowTitle_PushesAndPops(t *testing.T) {
	var out strings.Builder

	prev := TitleOut
	TitleOut = &out

	t.Cleanup(func() { TitleOut = prev })

	restore := SaveWindowTitle()

	assert.Equal(t, "\x1b[22;2t", out.String(), "the terminal's own title is saved before the program starts")

	restore()
	assert.Equal(t, "\x1b[22;2t\x1b[23;2t", out.String(), "and put back after it exits")
}

func TestWindowTitle(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "Show HN: a thing", "Show HN: a thing"},
		{"BEL would close the sequence", "safe\x07]0;evil", "safe]0;evil"},
		{"escape sequence", "safe\x1b[31mred", "safered"},
		{"newline would spill onto the screen", "first\nsecond", "first second"},
		{"whitespace collapses", "  spaced \t out  ", "spaced out"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, WindowTitle(tt.in))
		})
	}
}
