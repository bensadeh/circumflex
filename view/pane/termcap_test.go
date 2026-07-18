package pane

import (
	"testing"

	"github.com/bensadeh/circumflex/style"

	"github.com/stretchr/testify/assert"
)

func TestDetectStyledUnderline(t *testing.T) {
	clearTermEnv := func(t *testing.T) {
		t.Helper()
		t.Setenv("TERM", "")
		t.Setenv("TERM_PROGRAM", "")
		t.Setenv("VTE_VERSION", "")
	}

	t.Run("unknown terminals are queried for Smulx", func(t *testing.T) {
		clearTermEnv(t)
		t.Setenv("TERM", "xterm-256color")

		assert.NotNil(t, DetectStyledUnderline(), "the query is the only path to dashed underlines here")
	})

	t.Run("Apple Terminal is never queried", func(t *testing.T) {
		clearTermEnv(t)
		t.Setenv("TERM_PROGRAM", "Apple_Terminal")

		assert.Nil(t, DetectStyledUnderline(), "a DCS query would print as garbage in Terminal.app")
	})

	// The enabling cases run last: the flag is package-global and sticky.
	t.Run("iTerm2 enables dashed without a query", func(t *testing.T) {
		clearTermEnv(t)
		t.Setenv("TERM_PROGRAM", "iTerm.app")

		assert.Nil(t, DetectStyledUnderline())
		assert.Contains(t, style.ReaderLinkInert("x", "https://a"), "\x1b[4:5m")
	})
}
