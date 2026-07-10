package meta

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The package contract: for the same Data, a block's skeleton has exactly
// the rows of its rendered form, with the accent bar in the same column.
// This is what lets a view reserve the block's spot while fetching — nothing
// may move when the content arrives. Any block redesign has to keep this
// sweep green.
func TestSkeletonMatchesRender(t *testing.T) {
	linked := Data{
		URL:           "https://example.com/story",
		Domain:        "example.com",
		Author:        "alice",
		TimeAgo:       "2 hours ago",
		ID:            12345,
		Points:        100,
		CommentsCount: 42,
		NewComments:   3,
	}

	selfPost := Data{
		Author:        "bob",
		TimeAgo:       "1 hour ago",
		ID:            678,
		Points:        10,
		CommentsCount: 3,
	}

	withRootComment := selfPost
	withRootComment.RootComment = "A question for the community\nspanning two lines"

	blocks := map[string]Block{
		"comments linked":       CommentSection(linked),
		"comments self post":    CommentSection(selfPost),
		"comments root comment": CommentSection(withRootComment),
		"reader":                ReaderMode(linked),
		"url only":              ReaderModeURL("https://example.com/story", false),
	}

	for name, block := range blocks {
		for _, nerdFonts := range []bool{false, true} {
			for _, width := range []int{20, 60, 80} {
				rendered := block.Render(width)
				skeleton := block.Skeleton(width)

				assert.Equal(t, lipgloss.Height(rendered), lipgloss.Height(skeleton),
					"%s: skeleton height must match render at width %d (nerdfonts %v)", name, width, nerdFonts)

				for _, view := range []string{rendered, skeleton} {
					for line := range strings.SplitSeq(view, "\n") {
						assert.True(t, strings.HasPrefix(xansi.Strip(line), " "+bar),
							"%s: every row carries the accent bar in the same column, got %q", name, line)
					}
				}
			}
		}
	}
}

func TestSkeletonIsEmptyAndDimmed(t *testing.T) {
	skeleton := CommentSection(Data{Domain: "example.com", URL: "https://example.com"}).Skeleton(60)

	lines := strings.Split(skeleton, "\n")
	require.NotEmpty(t, lines)

	for _, line := range lines {
		assert.Contains(t, line, "\x1b[2m", "every skeleton row renders dimmed")
	}

	frameRunes := strings.NewReplacer(bar, "", " ", "", "\n", "")
	assert.Empty(t, frameRunes.Replace(xansi.Strip(skeleton)), "skeleton must hold no text")
}
