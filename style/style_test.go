package style

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bensadeh/circumflex/ansi"
)

func TestForegroundCode(t *testing.T) {
	t.Parallel()

	code := ForegroundCode(lipgloss.Red)

	require.NotEmpty(t, code)
	assert.True(t, strings.HasPrefix(code, "\x1b["), "should be a raw ANSI escape")
	assert.True(t, strings.HasSuffix(code, "m"), "should be a complete SGR sequence")

	assert.Empty(t, ForegroundCode(lipgloss.NoColor{}))
}

func TestRoundedBox_Label(t *testing.T) {
	t.Parallel()

	unlabeled := ansi.Strip(RoundedBox("some code", 17, ""))
	assert.Equal(t, "╭───────────────╮", strings.Split(unlabeled, "\n")[0], "no label keeps the plain rule")

	labeled := ansi.Strip(RoundedBox("some code", 17, "Go"))
	assert.Equal(t, "╭────────── Go ─╮", strings.Split(labeled, "\n")[0], "label right-aligns and the rule spans the same width")

	tooNarrow := ansi.Strip(RoundedBox("wide", 0, "VeryLongLanguageName"))
	assert.Equal(t, "╭──────╮", strings.Split(tooNarrow, "\n")[0], "an oversized label drops instead of breaking the frame")
}
