package style

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForegroundCode(t *testing.T) {
	t.Parallel()

	code := ForegroundCode(lipgloss.Red)

	require.NotEmpty(t, code)
	assert.True(t, strings.HasPrefix(code, "\x1b["), "should be a raw ANSI escape")
	assert.True(t, strings.HasSuffix(code, "m"), "should be a complete SGR sequence")

	assert.Empty(t, ForegroundCode(lipgloss.NoColor{}))
}
