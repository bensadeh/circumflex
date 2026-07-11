package headline

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

func TestHighlightDomain(t *testing.T) {
	t.Parallel()

	t.Run("empty domain returns reset only", func(t *testing.T) {
		t.Parallel()

		result := HighlightDomain("")
		assert.Equal(t, "\033[0m", result)
	})

	t.Run("non-empty domain contains domain text", func(t *testing.T) {
		t.Parallel()

		result := HighlightDomain("example.com")
		assert.Contains(t, result, "example.com")
		assert.True(t, strings.HasPrefix(result, "\033[0m"))
	})
}

func TestLabel_AppliesHighlightTypeStyles(t *testing.T) {
	t.Parallel()

	fg := lipgloss.Black
	bg := lipgloss.Yellow

	unselected := label("YC S20", fg, bg, Unselected)

	assert.NotEqual(t, unselected, label("YC S20", fg, bg, MarkAsRead),
		"mark-as-read should dim the label")
	assert.NotEqual(t, unselected, label("YC S20", fg, bg, HeadlineInCommentSection),
		"comment section headlines should embolden the label")
}
