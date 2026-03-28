package tree_test

import (
	"clx/comment"
	"clx/settings"
	"clx/tree"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getConfig() *settings.Config {
	return &settings.Config{
		CommentWidth:      110,
		IndentationSymbol: "▎",
	}
}

func TestPrintEmptyComments(t *testing.T) {
	t.Parallel()

	thread := &comment.Thread{Title: "Test", Author: "alice", Points: 10, CommentsCount: 0}
	result := tree.Print(thread, getConfig(), 120, 0)

	assert.NotEmpty(t, result)
}

func TestPrintSingleComment(t *testing.T) {
	t.Parallel()

	thread := &comment.Thread{
		Title: "Test", Author: "poster", Points: 10,
		Comments: []*comment.Comment{
			{ID: 1, Author: "bob", Content: "Hello world", Depth: 0, TimeAgo: "1h ago", Time: 50},
		},
	}
	result := tree.Print(thread, getConfig(), 120, 100)

	assert.Contains(t, result, "bob")
	assert.Contains(t, result, "Hello world")
}

func TestPrintCommentSeparator(t *testing.T) {
	t.Parallel()

	thread := &comment.Thread{
		Title: "Test", Author: "poster", Points: 10,
		Comments: []*comment.Comment{
			{ID: 1, Author: "bob", Content: "First", Depth: 0, TimeAgo: "2h ago", Time: 50},
			{ID: 2, Author: "carol", Content: "Second", Depth: 0, TimeAgo: "1h ago", Time: 60},
		},
	}
	result := tree.Print(thread, getConfig(), 120, 100)

	assert.Contains(t, result, "▁")
}

func TestPrintNestedComments(t *testing.T) {
	t.Parallel()

	thread := &comment.Thread{
		Title: "Test", Author: "poster", Points: 10,
		Comments: []*comment.Comment{
			{
				ID: 1, Author: "bob", Content: "Parent", Depth: 0, TimeAgo: "2h ago", Time: 50,
				Children: []*comment.Comment{
					{ID: 2, Author: "carol", Content: "Reply", Depth: 1, TimeAgo: "1h ago", Time: 60},
				},
			},
		},
	}
	result := tree.Print(thread, getConfig(), 120, 100)

	assert.Contains(t, result, "carol")
	assert.Contains(t, result, "Reply")
	assert.Contains(t, result, "▎")
}

func TestPrintModLabel(t *testing.T) {
	t.Parallel()

	thread := &comment.Thread{
		Title: "Test", Author: "poster", Points: 10,
		Comments: []*comment.Comment{
			{ID: 1, Author: "dang", Content: "Mod says hi", Depth: 0, TimeAgo: "1h ago", Time: 50},
		},
	}
	result := tree.Print(thread, getConfig(), 120, 100)

	assert.Contains(t, result, "mod")
}

func TestPrintOPLabel(t *testing.T) {
	t.Parallel()

	thread := &comment.Thread{
		Title: "Test", Author: "poster", Points: 10,
		Comments: []*comment.Comment{
			{ID: 1, Author: "poster", Content: "OP here", Depth: 0, TimeAgo: "1h ago", Time: 50},
		},
	}
	result := tree.Print(thread, getConfig(), 120, 100)

	assert.Contains(t, result, "OP")
}

func TestPrintNewCommentDot(t *testing.T) {
	t.Parallel()

	thread := &comment.Thread{
		Title: "Test", Author: "poster", Points: 10,
		Comments: []*comment.Comment{
			{ID: 1, Author: "bob", Content: "New comment", Depth: 0, TimeAgo: "1m ago", Time: 200},
		},
	}
	result := tree.Print(thread, getConfig(), 120, 100) // lastVisited=100 < Time=200

	assert.Contains(t, result, "●")
}

func TestPrintDeletedCommentSkipped(t *testing.T) {
	t.Parallel()

	thread := &comment.Thread{
		Title: "Test", Author: "poster", Points: 10,
		Comments: []*comment.Comment{
			{ID: 1, Author: "deleted_user", Content: "[deleted]", Depth: 0, TimeAgo: "1h ago", Time: 50},
		},
	}
	result := tree.Print(thread, getConfig(), 120, 100)

	assert.NotContains(t, result, "deleted_user")
}

func TestPrintDeletedCommentWithReplies(t *testing.T) {
	t.Parallel()

	thread := &comment.Thread{
		Title: "Test", Author: "poster", Points: 10,
		Comments: []*comment.Comment{
			{
				ID: 1, Author: "deleted_user", Content: "[deleted]", Depth: 0, TimeAgo: "1h ago", Time: 50,
				Children: []*comment.Comment{
					{ID: 2, Author: "carol", Content: "Reply to deleted", Depth: 1, TimeAgo: "30m ago", Time: 60},
				},
			},
		},
	}
	result := tree.Print(thread, getConfig(), 120, 100)

	assert.Contains(t, result, "[deleted]")
}
