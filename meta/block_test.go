package meta

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The package contract, enforced across every variant, width parity, and
// nerd-font setting:
//
//  1. A block's skeleton has exactly the rows of its rendered form, with the
//     accent bar in the same column — a view can reserve the block's spot
//     while fetching and nothing moves when the content arrives.
//  2. Blocks carry no left margin: every row opens with the accent bar. The
//     hosting view supplies the margin, so the block can never disagree with
//     the text column it heads about where the margin ends.
//  3. No row extends past width-rightInset, and the right-aligned rows land
//     exactly there. The right edge depends on nothing but the width — a
//     change to the frame, the insets, or the hosting margins that moves it
//     fails here.
//
// Any block redesign has to keep this sweep green.
func TestBlockGeometryContract(t *testing.T) {
	for _, nerdFonts := range []bool{false, true} {
		linked := Data{
			URL:           "https://example.com/story",
			Domain:        "example.com",
			Author:        "alice",
			TimeAgo:       "2 hours ago",
			ID:            12345,
			Points:        100,
			CommentsCount: 42,
			NewComments:   3,
			NerdFonts:     nerdFonts,
		}

		selfPost := Data{
			Author:        "bob",
			TimeAgo:       "1 hour ago",
			ID:            678,
			Points:        10,
			CommentsCount: 3,
			NerdFonts:     nerdFonts,
		}

		for _, width := range []int{20, 21, 60, 61, 80} {
			// Callers wrap pre-rendered content at ContentWidth; the fixture
			// honors the same contract.
			withRootComment := selfPost
			withRootComment.RootComment = lipgloss.Wrap(
				"A question for the community spanning multiple lines once wrapped", ContentWidth(width), "")

			// A submission can carry both a link and text; this is the layout
			// with every part present, including the rule between them.
			linkedWithText := linked
			linkedWithText.RootComment = withRootComment.RootComment

			// hasColumns marks the variants with right-aligned ID/points rows,
			// which must reach the right edge exactly.
			blocks := map[string]struct {
				block      Block
				hasColumns bool
			}{
				"comments linked":           {CommentSection(linked), true},
				"comments self post":        {CommentSection(selfPost), true},
				"comments root comment":     {CommentSection(withRootComment), true},
				"comments linked with text": {CommentSection(linkedWithText), true},
				"reader":                    {ReaderMode(linked), true},
				"url only":                  {ReaderModeURL("https://example.com/story", nerdFonts), false},
			}

			for name, tc := range blocks {
				label := fmt.Sprintf("%s at width %d (nerdfonts %v)", name, width, nerdFonts)

				rendered := strings.Split(tc.block.Render(width), "\n")
				skeleton := strings.Split(tc.block.Skeleton(width), "\n")

				require.Len(t, skeleton, len(rendered),
					"%s: skeleton height must match render", label)

				edge := 0

				for i := range rendered {
					assert.True(t, strings.HasPrefix(xansi.Strip(rendered[i]), bar),
						"%s: row %d must open with the accent bar, got %q", label, i, rendered[i])
					assert.True(t, strings.HasPrefix(xansi.Strip(skeleton[i]), bar),
						"%s: skeleton row %d must carry the accent bar in the same column, got %q", label, i, skeleton[i])

					rowWidth := xansi.StringWidth(rendered[i])
					assert.LessOrEqual(t, rowWidth, width-rightInset,
						"%s: row %d extends past the right inset: %q", label, i, rendered[i])
					edge = max(edge, rowWidth)
				}

				if tc.hasColumns {
					assert.Equal(t, width-rightInset, edge,
						"%s: the right-aligned rows must end exactly at width-rightInset", label)
				}
			}
		}
	}
}

// The URL is the block's last row. When the submission also carries its own
// text, a rule sits between the text and the link so the link can't read as
// the text's closing paragraph; with no text there is nothing to confuse and
// no rule appears.
func TestURLIsTheFooter(t *testing.T) {
	linked := Data{URL: "https://example.com/story", Domain: "example.com"}

	withText := linked
	withText.RootComment = "A question for the community"

	rows := strings.Split(xansi.Strip(CommentSection(withText).Render(60)), "\n")
	require.GreaterOrEqual(t, len(rows), 3)

	last := rows[len(rows)-1]
	rule := rows[len(rows)-2]

	assert.Contains(t, last, "https://example.com/story", "the URL must be the last row")
	assert.Equal(t, bar+" "+strings.Repeat("─", ContentWidth(60)), rule,
		"a rule must separate the submission text from the URL")

	assert.NotContains(t, CommentSection(linked).Render(60), "─",
		"no rule without submission text")
}

// A URL wider than the content width shortens to a single-character
// ellipsis, never wrapping or spilling past the block's edge.
func TestURLTruncatesWithEllipsis(t *testing.T) {
	url := "https://example.com/a/very/long/path/that/cannot/possibly/fit"
	rendered := CommentSection(Data{URL: url, Domain: "example.com"}).Render(30)

	rows := strings.Split(xansi.Strip(rendered), "\n")
	last := rows[len(rows)-1]

	assert.True(t, strings.HasSuffix(last, "…"), "truncated URL must end in a single ellipsis, got %q", last)
	assert.Equal(t, 30-rightInset, xansi.StringWidth(last), "truncated URL must fill exactly to the right inset")
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
