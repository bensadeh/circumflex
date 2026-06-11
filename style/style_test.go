package style

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// foregroundCode extracts the escape sequence by rendering a marker through
// lipgloss and slicing — this pins that trick against lipgloss render changes.
func TestForegroundCode(t *testing.T) {
	t.Parallel()

	code := foregroundCode(lipgloss.Red)

	require.NotEmpty(t, code)
	assert.True(t, strings.HasPrefix(code, "\x1b["), "should be a raw ANSI escape")
	assert.True(t, strings.HasSuffix(code, "m"), "should be a complete SGR sequence")
	assert.NotContains(t, code, "\xff", "marker must not leak into the result")

	assert.Empty(t, foregroundCode(lipgloss.NoColor{}))
}
