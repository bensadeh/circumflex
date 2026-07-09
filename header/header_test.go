package header

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnderline_MemorialTinting(t *testing.T) {
	t.Cleanup(func() { SetMemorial(false) })

	SetMemorial(false)

	plain := Underline(10)
	assert.Equal(t, strings.Repeat("‾", 10), plain, "inactive underline should be unstyled")
	assert.NotContains(t, plain, "\x1b[", "inactive underline should carry no ANSI escape")

	SetMemorial(true)

	tinted := Underline(10)
	assert.Contains(t, tinted, "\x1b[", "active underline should be styled with an ANSI escape")
	assert.Contains(t, tinted, strings.Repeat("═", 10), "active underline should render as a double line")
	assert.NotEqual(t, plain, tinted, "active and inactive underlines should differ")
}
