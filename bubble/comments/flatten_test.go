package comments

import (
	"clx/comment"
	"clx/settings"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newComment(id int, author, content string, children ...*comment.Comment) *comment.Comment {
	return &comment.Comment{
		ID:       id,
		Author:   author,
		Content:  content,
		Time:     int64(id * 100),
		TimeAgo:  "1 hour ago",
		Children: children,
	}
}

func newThread(comments ...*comment.Comment) *comment.Thread {
	return &comment.Thread{
		ID:       1,
		Title:    "test",
		Author:   "op",
		Comments: comments,
	}
}

func expandAll(flat []FlatComment) {
	for i := range flat {
		flat[i].Collapsed = false
	}
}

// --- flatten: DFS ordering invariant ---

func TestFlatten_DFSOrder(t *testing.T) {
	t.Parallel()

	//   A (depth 0)
	//     B (depth 1)
	//       C (depth 2)
	//     D (depth 1)
	//   E (depth 0)
	thread := newThread(
		newComment(1, "alice", "A",
			newComment(2, "bob", "B",
				newComment(3, "charlie", "C"),
			),
			newComment(4, "dave", "D"),
		),
		newComment(5, "eve", "E"),
	)

	flat := flatten(thread)

	// DFS invariant: depth increases by exactly 1 when it increases.
	for i := 1; i < len(flat); i++ {
		if flat[i].Depth > flat[i-1].Depth {
			assert.Equal(t, flat[i-1].Depth+1, flat[i].Depth,
				"depth must increase by exactly 1 at index %d", i)
		}
	}

	require.Len(t, flat, 5)
	assert.Equal(t, 1, flat[0].Comment.ID)
	assert.Equal(t, 2, flat[1].Comment.ID)
	assert.Equal(t, 3, flat[2].Comment.ID)
	assert.Equal(t, 4, flat[3].Comment.ID)
	assert.Equal(t, 5, flat[4].Comment.ID)
}

func TestFlatten_Depths(t *testing.T) {
	t.Parallel()

	thread := newThread(
		newComment(1, "alice", "A",
			newComment(2, "bob", "B",
				newComment(3, "charlie", "C"),
			),
		),
	)

	flat := flatten(thread)

	require.Len(t, flat, 3)
	assert.Equal(t, 0, flat[0].Depth)
	assert.Equal(t, 1, flat[1].Depth)
	assert.Equal(t, 2, flat[2].Depth)
}

func TestFlatten_DescendantCount(t *testing.T) {
	t.Parallel()

	thread := newThread(
		newComment(1, "alice", "A",
			newComment(2, "bob", "B",
				newComment(3, "charlie", "C"),
			),
			newComment(4, "dave", "D"),
		),
	)

	flat := flatten(thread)

	require.Len(t, flat, 4)
	assert.Equal(t, 3, flat[0].DescendantCount) // A → B, C, D
	assert.Equal(t, 1, flat[1].DescendantCount) // B → C
	assert.Equal(t, 0, flat[2].DescendantCount) // C leaf
	assert.Equal(t, 0, flat[3].DescendantCount) // D leaf
}

func TestFlatten_SkipsDeletedLeaves(t *testing.T) {
	t.Parallel()

	thread := newThread(
		newComment(1, "alice", "A",
			newComment(2, "bob", "[deleted]"),
			newComment(3, "charlie", "C"),
		),
	)

	flat := flatten(thread)

	require.Len(t, flat, 2)
	assert.Equal(t, 1, flat[0].Comment.ID)
	assert.Equal(t, 3, flat[1].Comment.ID)
}

func TestFlatten_KeepsDeletedWithChildren(t *testing.T) {
	t.Parallel()

	thread := newThread(
		newComment(1, "alice", "[deleted]",
			newComment(2, "bob", "reply"),
		),
	)

	flat := flatten(thread)

	require.Len(t, flat, 2)
	assert.Equal(t, "[deleted]", flat[0].Comment.Content)
	assert.Equal(t, "reply", flat[1].Comment.Content)
}

func TestFlatten_TopLevelAuthor(t *testing.T) {
	t.Parallel()

	thread := newThread(
		newComment(1, "alice", "A",
			newComment(2, "bob", "B",
				newComment(3, "charlie", "C"),
			),
		),
		newComment(4, "dave", "D",
			newComment(5, "eve", "E"),
		),
	)

	flat := flatten(thread)

	require.Len(t, flat, 5)
	assert.Empty(t, flat[0].TopLevelAuthor)          // A at depth 0
	assert.Equal(t, "alice", flat[1].TopLevelAuthor) // B inherits A's author
	assert.Equal(t, "alice", flat[2].TopLevelAuthor) // C inherits A's author
	assert.Empty(t, flat[3].TopLevelAuthor)          // D at depth 0
	assert.Equal(t, "dave", flat[4].TopLevelAuthor)  // E inherits D's author
}

func TestFlatten_InitialCollapseState(t *testing.T) {
	t.Parallel()

	thread := newThread(
		newComment(1, "alice", "A",
			newComment(2, "bob", "B"),
		),
		newComment(3, "charlie", "C"),
	)

	flat := flatten(thread)

	require.Len(t, flat, 3)
	assert.True(t, flat[0].Collapsed, "top-level with children starts collapsed")
	assert.False(t, flat[1].Collapsed, "non-top-level never starts collapsed")
	assert.False(t, flat[2].Collapsed, "top-level without children not collapsed")
}

func TestFlatten_EmptyThread(t *testing.T) {
	t.Parallel()

	flat := flatten(newThread())

	assert.Empty(t, flat)
}

// --- computeVisible ---

func TestComputeVisible_AllExpanded(t *testing.T) {
	t.Parallel()

	flat := []FlatComment{
		{Depth: 0, DescendantCount: 2},
		{Depth: 1, DescendantCount: 1},
		{Depth: 2, DescendantCount: 0},
	}

	assert.Equal(t, []int{0, 1, 2}, computeVisible(flat))
}

func TestComputeVisible_CollapsedSkipsDescendants(t *testing.T) {
	t.Parallel()

	flat := []FlatComment{
		{Depth: 0, Collapsed: true, DescendantCount: 3}, // A
		{Depth: 1, DescendantCount: 1},                  // B (hidden)
		{Depth: 2, DescendantCount: 0},                  // C (hidden)
		{Depth: 1, DescendantCount: 0},                  // D (hidden)
		{Depth: 0, DescendantCount: 0},                  // E
	}

	assert.Equal(t, []int{0, 4}, computeVisible(flat))
}

func TestComputeVisible_NestedCollapse(t *testing.T) {
	t.Parallel()

	flat := []FlatComment{
		{Depth: 0, Collapsed: true, DescendantCount: 2}, // A (collapsed)
		{Depth: 1, Collapsed: true, DescendantCount: 1}, // B (collapsed, but hidden by A)
		{Depth: 2, DescendantCount: 0},                  // C (hidden)
		{Depth: 0, DescendantCount: 0},                  // D
	}

	assert.Equal(t, []int{0, 3}, computeVisible(flat))
}

func TestComputeVisible_MidTreeCollapse(t *testing.T) {
	t.Parallel()

	flat := []FlatComment{
		{Depth: 0, DescendantCount: 3},                  // A
		{Depth: 1, Collapsed: true, DescendantCount: 1}, // B (collapsed)
		{Depth: 2, DescendantCount: 0},                  // C (hidden by B)
		{Depth: 1, DescendantCount: 0},                  // D
	}

	assert.Equal(t, []int{0, 1, 3}, computeVisible(flat))
}

func TestComputeVisible_Empty(t *testing.T) {
	t.Parallel()

	assert.Empty(t, computeVisible(nil))
}

func TestComputeVisible_CollapsedLeafNoEffect(t *testing.T) {
	t.Parallel()

	flat := []FlatComment{
		{Depth: 0, Collapsed: true, DescendantCount: 0},
		{Depth: 0, DescendantCount: 0},
	}

	assert.Equal(t, []int{0, 1}, computeVisible(flat),
		"collapsed node with no descendants should not hide anything")
}

// --- renderFromFlat contract tests ---

func defaultRenderContext() renderContext {
	return renderContext{
		originalPoster: "op",
		firstCommentID: 1,
		config:         settings.Default(),
		screenWidth:    80,
		viewportHeight: 40,
	}
}

func TestRenderFromFlat_LineMetricsMonotonic(t *testing.T) {
	t.Parallel()

	thread := newThread(
		newComment(1, "alice", "First comment"),
		newComment(2, "bob", "Second comment"),
		newComment(3, "charlie", "Third comment"),
	)

	flat := flatten(thread)
	expandAll(flat)

	visible := computeVisible(flat)
	rc := defaultRenderContext()
	_, _, metrics := renderFromFlat(rc, flat, visible, prerenderComments(rc, flat))

	prevStart := -1

	for _, vi := range visible {
		lm := metrics[vi]

		assert.Greater(t, lm.StartLine, prevStart,
			"StartLine must be strictly increasing (flat index %d)", vi)

		prevStart = lm.StartLine
	}
}

func TestRenderFromFlat_LineCountMatchesContent(t *testing.T) {
	t.Parallel()

	thread := newThread(
		newComment(1, "alice", "Hello world"),
		newComment(2, "bob", "Another comment"),
	)

	flat := flatten(thread)
	expandAll(flat)

	visible := computeVisible(flat)
	rc := defaultRenderContext()
	content, contentLines, metrics := renderFromFlat(rc, flat, visible, prerenderComments(rc, flat))

	// Total newlines = content lines + bottom padding.
	totalNewlines := strings.Count(content, "\n")
	assert.Equal(t, contentLines+rc.viewportHeight, totalNewlines)

	// Every visible comment must have a positive line count.
	for _, vi := range visible {
		assert.Positive(t, metrics[vi].LineCount,
			"visible comment must have positive LineCount (flat index %d)", vi)
	}

	// Last comment must end within content bounds.
	last := metrics[visible[len(visible)-1]]
	assert.LessOrEqual(t, last.StartLine+last.LineCount, contentLines)
}

func TestRenderFromFlat_CollapsedShowsFoldIndicator(t *testing.T) {
	t.Parallel()

	thread := newThread(
		newComment(1, "alice", "parent",
			newComment(2, "bob", "child"),
		),
	)

	flat := flatten(thread)
	// flat[0] starts collapsed (depth 0, has child)

	visible := computeVisible(flat)
	rc := defaultRenderContext()
	content, _, _ := renderFromFlat(rc, flat, visible, prerenderComments(rc, flat))

	assert.Contains(t, content, "1 reply hidden")
}

func TestRenderFromFlat_NonVisibleMetricsAreZero(t *testing.T) {
	t.Parallel()

	thread := newThread(
		newComment(1, "alice", "parent",
			newComment(2, "bob", "hidden child"),
		),
	)

	flat := flatten(thread)

	visible := computeVisible(flat)
	require.Len(t, visible, 1)

	rc := defaultRenderContext()
	_, _, metrics := renderFromFlat(rc, flat, visible, prerenderComments(rc, flat))

	assert.Equal(t, LineMetrics{}, metrics[1])
}
