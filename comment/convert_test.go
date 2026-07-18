package comment

import (
	"testing"

	"github.com/bensadeh/circumflex/hn"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func node(id int, author, content string, children ...*hn.CommentNode) *hn.CommentNode {
	return &hn.CommentNode{
		ID:       id,
		Author:   author,
		Content:  content,
		Time:     int64(id * 100),
		Children: children,
	}
}

func TestToThread_PrunesRemovedLeaves(t *testing.T) {
	t.Parallel()

	thread := ToThread(&hn.CommentTree{Comments: []*hn.CommentNode{
		node(1, "alice", "A",
			node(2, "bob", "[deleted]"),
			node(3, "charlie", "C"),
			node(4, "dave", "[flagged]"),
			node(5, "erin", "[delayed]"),
		),
	}})

	require.Len(t, thread.Comments, 1)
	require.Len(t, thread.Comments[0].Children, 1)
	assert.Equal(t, 3, thread.Comments[0].Children[0].ID)
}

func TestToThread_PruneIsTransitive(t *testing.T) {
	t.Parallel()

	// A removed comment whose replies were all pruned goes with them.
	thread := ToThread(&hn.CommentTree{Comments: []*hn.CommentNode{
		node(1, "", "[deleted]",
			node(2, "", "[deleted]"),
			node(3, "", "[flagged]"),
		),
		node(4, "alice", "A"),
	}})

	require.Len(t, thread.Comments, 1)
	assert.Equal(t, 4, thread.Comments[0].ID)
}

func TestToThread_KeepsRemovedAnchors(t *testing.T) {
	t.Parallel()

	thread := ToThread(&hn.CommentTree{Comments: []*hn.CommentNode{
		node(1, "", "[deleted]",
			node(2, "bob", "reply"),
		),
	}})

	require.Len(t, thread.Comments, 1)
	assert.Equal(t, "[deleted]", thread.Comments[0].Content)
	require.Len(t, thread.Comments[0].Children, 1)
	assert.Equal(t, "reply", thread.Comments[0].Children[0].Content)
}

// A pruned first comment must not leave its ID behind as the separator
// anchor: the first rendered comment is the first one in the pruned tree.
func TestToThread_FirstCommentSurvivesPrune(t *testing.T) {
	t.Parallel()

	thread := ToThread(&hn.CommentTree{Comments: []*hn.CommentNode{
		node(1, "", "[deleted]"),
		node(2, "alice", "A"),
	}})

	assert.Equal(t, 2, FirstCommentID(thread.Comments))
}

// The "new comments" badge counts rendered comments only: a fresh removed
// leaf is pruned before counting.
func TestNewCommentsCount_IgnoresPrunedComments(t *testing.T) {
	t.Parallel()

	thread := ToThread(&hn.CommentTree{Comments: []*hn.CommentNode{
		node(1, "alice", "A"),
		node(2, "bob", "[deleted]"),
	}})

	assert.Equal(t, 1, NewCommentsCount(thread, 0))
}
