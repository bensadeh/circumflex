package list

import (
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/categories"
	"github.com/bensadeh/circumflex/history"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The open story renders on a muted bright-black bar — a dimmed version of
// the browsing highlight, marking the J/K reading position — while the other
// stories dim.
func TestRenderItem_OpenStoryShowsReadingMarker(t *testing.T) {
	cat, err := categories.New("top,best,ask,show")
	require.NoError(t, err)

	m := New(settings.Default(), cat, history.NewMockHistory())
	m.SetItems(categories.Top, []*hn.Story{
		{ID: 1, Title: "First item", Points: 100, Author: "alice", Domain: "example.com"},
		{ID: 2, Title: "Second item", Points: 200, Author: "bob", Domain: "test.com"},
	})
	m.Resize(123, 21)

	f := Frame{Wide: true, DetailOpen: true}
	require.True(t, m.dimmed(f))
	require.True(t, m.storyOpen(f))

	var open strings.Builder

	m.renderItem(&open, m.Index(), m.SelectedItem(), f)
	assert.NotContains(t, open.String(), "\x1b[7m", "open story should not use the full browsing highlight")
	assert.Contains(t, open.String(), "\x1b[100m", "open story should render on a bright-black bar")

	var other strings.Builder

	m.renderItem(&other, m.Index()+1, m.VisibleItems()[m.Index()+1], f)
	assert.Contains(t, other.String(), "\x1b[3;2m", "other stories should dim")
}

// Highlighted tokens (Show HN:, years, YC tags) reset the row's style and
// restore it afterwards; on the open story that restore must re-apply the
// bar's background, not Selected's reverse video.
func TestRenderItem_OpenStoryBarSurvivesHighlightedTokens(t *testing.T) {
	cat, err := categories.New("top,best,ask,show")
	require.NoError(t, err)

	m := New(settings.Default(), cat, history.NewMockHistory())
	m.SetItems(categories.Top, []*hn.Story{
		{ID: 1, Title: "Show HN: Time capsule (2019) (YC W19)", Points: 100, Author: "alice"},
	})
	m.Resize(123, 21)

	f := Frame{Wide: true, DetailOpen: true}

	var open strings.Builder

	m.renderItem(&open, m.Index(), m.SelectedItem(), f)
	rendered := open.String()

	assert.NotContains(t, rendered, "\x1b[7m", "token restores must not fall back to reverse video")

	title, _, _ := strings.Cut(rendered, "\n")
	// The empty domain still appends a bare reset; drop it so the indexes
	// below see the last token's reset, not the domain's.
	title = strings.TrimSuffix(title, " \x1b[0m")

	lastReset := strings.LastIndex(title, "\x1b[0m")
	require.GreaterOrEqual(t, lastReset, 0, "highlighted tokens should reset the row style")

	lastBar := strings.LastIndex(title, "\x1b[100m")
	assert.Greater(t, lastBar, lastReset, "the bar should be restored after the last token reset")
}
