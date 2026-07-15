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
//  1. A block's skeleton has exactly the rows of its rendered form, wrapped
//     in the same frame — a view can reserve the block's spot while fetching
//     and nothing moves when the content arrives.
//  2. Blocks carry no left margin. The hosting view supplies the margin, so
//     the block can never disagree with the text column it heads about where
//     the margin ends.
//  3. Every row spans exactly width cells: the frame's opening rule, its
//     body rows, and its closing rule all reach that edge and no further —
//     the same right edge the help panels drawn over the column share. A
//     change to the frame or the hosting margins that pushes a row past it
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
			Points:        100,
			CommentsCount: 33,
			NewComments:   5,
			NerdFonts:     nerdFonts,
		}

		selfPost := Data{
			Author:        "bob",
			TimeAgo:       "1 hour ago",
			Points:        10,
			CommentsCount: 2,
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

			blocks := map[string]Block{
				"comments linked":           CommentSection(linked),
				"comments self post":        CommentSection(selfPost),
				"comments root comment":     CommentSection(withRootComment),
				"comments linked with text": CommentSection(linkedWithText),
				"reader":                    ReaderMode(linked),
				"url only":                  ReaderModeURL("https://example.com/story"),
			}

			for name, block := range blocks {
				label := fmt.Sprintf("%s at width %d (nerdfonts %v)", name, width, nerdFonts)

				rendered := strings.Split(block.Render(width), "\n")
				skeleton := strings.Split(block.Skeleton(width), "\n")

				require.Len(t, skeleton, len(rendered),
					"%s: skeleton height must match render", label)

				edge := width

				opening := "╭" + strings.Repeat("─", edge-2) + "╮"
				closing := "╰" + strings.Repeat("─", edge-2) + "╯"

				assert.Equal(t, closing, xansi.Strip(rendered[len(rendered)-1]),
					"%s: the closing rule must be the last row, spanning the full width", label)
				assert.Equal(t, closing, xansi.Strip(skeleton[len(skeleton)-1]),
					"%s: the skeleton must close with the same rule", label)
				assert.Equal(t, opening, xansi.Strip(skeleton[0]),
					"%s: the skeleton's opening rule carries no title", label)

				for i, row := range skeleton[1 : len(skeleton)-1] {
					assert.Equal(t, "│"+strings.Repeat(" ", edge-2)+"│", xansi.Strip(row),
						"%s: skeleton row %d must be a blank framed row", label, i+1)
				}

				for i := range rendered {
					assert.Equal(t, edge, xansi.StringWidth(rendered[i]),
						"%s: row %d must reach the frame's edge exactly: %q", label, i, rendered[i])
				}

				top := xansi.Strip(rendered[0])
				assert.True(t, strings.HasPrefix(top, "╭") && strings.HasSuffix(top, "╮"),
					"%s: the opening rule must span the block: %q", label, top)

				for i, row := range rendered[1 : len(rendered)-1] {
					s := xansi.Strip(row)
					assert.True(t, strings.HasPrefix(s, "│") && strings.HasSuffix(s, "│"),
						"%s: body row %d must sit inside the side borders: %q", label, i+1, s)
				}
			}
		}
	}
}

// The opening rule doubles as the block's header: the byline sits in it the
// way a help-panel title does, and the stat labels — comment count, then
// score — close it against the right corner with a rule segment between
// them. When the rule can't carry everything, the labels shed from the left,
// the count before the score, then the byline — the frame never gives up its
// own corners.
func TestOpeningRuleCarriesBylineAndStats(t *testing.T) {
	d := Data{
		URL:           "https://example.com/story",
		Domain:        "example.com",
		Author:        "alice",
		TimeAgo:       "2 hours ago",
		Points:        100,
		CommentsCount: 45,
	}

	top := func(width int) string {
		return strings.Split(xansi.Strip(CommentSection(d).Render(width)), "\n")[0]
	}

	wide := top(60)
	assert.True(t, strings.HasPrefix(wide, "╭── by alice 2 hours ago ─"),
		"the byline must open the rule: %q", wide)
	assert.True(t, strings.HasSuffix(wide, "─ 45 comments ─ 100 points ──╮"),
		"the comment count and score must close the rule right-aligned: %q", wide)

	mid := top(45)
	assert.Contains(t, mid, "points", "the score stays when only the count is out of room: %q", mid)
	assert.NotContains(t, mid, "comments", "the comment count is the first thing to go: %q", mid)

	tight := top(30)
	assert.Contains(t, tight, "by alice", "the byline stays when the labels are out of room: %q", tight)
	assert.NotContains(t, tight, "points", "the score goes before the byline: %q", tight)

	narrow := top(20)
	assert.Equal(t, "╭"+strings.Repeat("─", 18)+"╮", narrow,
		"a rule too narrow for any text stays plain")
}

// The new-comments count rides the tally in parentheses; without new
// comments there is no parenthetical at all.
func TestCommentTallyCarriesNewComments(t *testing.T) {
	d := Data{Author: "alice", TimeAgo: "2 hours ago", Points: 100, CommentsCount: 45}

	top := strings.Split(xansi.Strip(CommentSection(d).Render(80)), "\n")[0]
	assert.Contains(t, top, "─ 45 comments ─ ", "no parenthetical without new comments: %q", top)

	d.NewComments = 5
	top = strings.Split(xansi.Strip(CommentSection(d).Render(80)), "\n")[0]
	assert.Contains(t, top, "─ 45 comments (5 new) ─ ", "new comments join the tally: %q", top)
}

// The URL is the block's last row before the closing rule. When the
// submission also carries its own text, a light rule sits between the text
// and the link so the link can't read as the text's closing paragraph; with
// no text there is nothing to confuse and no light rule appears.
func TestURLIsTheFooter(t *testing.T) {
	linked := Data{URL: "https://example.com/story", Domain: "example.com"}

	withText := linked
	withText.RootComment = "A question for the community"

	rows := strings.Split(xansi.Strip(CommentSection(withText).Render(60)), "\n")
	require.GreaterOrEqual(t, len(rows), 4)

	url := rows[len(rows)-2]
	rule := rows[len(rows)-3]

	assert.Contains(t, url, "example.com/story", "the URL must be the last row before the closing rule")
	assert.NotContains(t, url, "https://", "the scheme is stripped from the display")
	assert.Contains(t, rule, strings.Repeat("─", ContentWidth(60)),
		"a rule must separate the submission text from the URL")

	assert.Len(t, strings.Split(CommentSection(linked).Render(60), "\n"), 3,
		"no rule row without submission text — just the URL inside the frame")

	assert.Contains(t, CommentSection(withText).Render(60), "https://example.com/story",
		"the hyperlink target keeps the full URL")
}

// A URL wider than the content width shortens to a single-character
// ellipsis, never wrapping or spilling past the frame.
func TestURLTruncatesWithEllipsis(t *testing.T) {
	url := "https://example.com/a/very/long/path/that/cannot/possibly/fit"
	rendered := CommentSection(Data{URL: url, Domain: "example.com"}).Render(30)

	rows := strings.Split(xansi.Strip(rendered), "\n")
	require.Len(t, rows, 3)
	urlRow := rows[1]

	assert.True(t, strings.HasSuffix(urlRow, "… │"),
		"truncated URL must end in a single ellipsis against the frame, got %q", urlRow)
	assert.Equal(t, 30, xansi.StringWidth(urlRow), "the URL row must fill the frame exactly")
}

func TestSkeletonIsEmptyAndDimmed(t *testing.T) {
	skeleton := CommentSection(Data{Domain: "example.com", URL: "https://example.com"}).Skeleton(60)

	lines := strings.Split(skeleton, "\n")
	require.NotEmpty(t, lines)

	assert.Contains(t, lines[len(lines)-1], "\x1b[2m", "the frame renders dimmed")

	frameRunes := strings.NewReplacer("╭", "", "╮", "", "╰", "", "╯", "", "│", "", "─", "", " ", "", "\n", "")
	assert.Empty(t, frameRunes.Replace(xansi.Strip(skeleton)), "skeleton must hold no text")
}
