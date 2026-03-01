package tree_test

import (
	"encoding/json"
	"testing"

	"clx/item"
	"clx/settings"
	"clx/tree"

	"github.com/stretchr/testify/assert"
)

func unmarshal(data []byte) *item.Item {
	root := new(item.Item)
	_ = json.Unmarshal(data, &root)

	return root
}

func getConfig() *settings.Config {
	return &settings.Config{
		CommentWidth:      110,
		IndentationSymbol: "▎",
	}
}

func TestPrintEmptyComments(t *testing.T) {
	t.Parallel()

	comments := &item.Item{Title: "Test", User: "alice", Points: 10, CommentsCount: 0}
	result := tree.Print(comments, getConfig(), 120, 0)

	assert.NotEmpty(t, result)
}

func TestPrintSingleComment(t *testing.T) {
	t.Parallel()

	comments := &item.Item{
		Title: "Test", User: "poster", Points: 10,
		Comments: []*item.Item{
			{ID: 1, User: "bob", Content: "Hello world", Level: 0, TimeAgo: "1h ago", Time: 50},
		},
	}
	result := tree.Print(comments, getConfig(), 120, 100)

	assert.Contains(t, result, "bob")
	assert.Contains(t, result, "Hello world")
}

func TestPrintCommentSeparator(t *testing.T) {
	t.Parallel()

	comments := &item.Item{
		Title: "Test", User: "poster", Points: 10,
		Comments: []*item.Item{
			{ID: 1, User: "bob", Content: "First", Level: 0, TimeAgo: "2h ago", Time: 50},
			{ID: 2, User: "carol", Content: "Second", Level: 0, TimeAgo: "1h ago", Time: 60},
		},
	}
	result := tree.Print(comments, getConfig(), 120, 100)

	assert.Contains(t, result, "▁")
}

func TestPrintNestedComments(t *testing.T) {
	t.Parallel()

	comments := &item.Item{
		Title: "Test", User: "poster", Points: 10,
		Comments: []*item.Item{
			{
				ID: 1, User: "bob", Content: "Parent", Level: 0, TimeAgo: "2h ago", Time: 50,
				Comments: []*item.Item{
					{ID: 2, User: "carol", Content: "Reply", Level: 1, TimeAgo: "1h ago", Time: 60},
				},
			},
		},
	}
	result := tree.Print(comments, getConfig(), 120, 100)

	assert.Contains(t, result, "carol")
	assert.Contains(t, result, "Reply")
	assert.Contains(t, result, "▎")
}

func TestPrintModLabel(t *testing.T) {
	t.Parallel()

	comments := &item.Item{
		Title: "Test", User: "poster", Points: 10,
		Comments: []*item.Item{
			{ID: 1, User: "dang", Content: "Mod says hi", Level: 0, TimeAgo: "1h ago", Time: 50},
		},
	}
	result := tree.Print(comments, getConfig(), 120, 100)

	assert.Contains(t, result, "mod")
}

func TestPrintOPLabel(t *testing.T) {
	t.Parallel()

	comments := &item.Item{
		Title: "Test", User: "poster", Points: 10,
		Comments: []*item.Item{
			{ID: 1, User: "poster", Content: "OP here", Level: 0, TimeAgo: "1h ago", Time: 50},
		},
	}
	result := tree.Print(comments, getConfig(), 120, 100)

	assert.Contains(t, result, "OP")
}

func TestPrintNewCommentDot(t *testing.T) {
	t.Parallel()

	comments := &item.Item{
		Title: "Test", User: "poster", Points: 10,
		Comments: []*item.Item{
			{ID: 1, User: "bob", Content: "New comment", Level: 0, TimeAgo: "1m ago", Time: 200},
		},
	}
	result := tree.Print(comments, getConfig(), 120, 100) // lastVisited=100 < Time=200

	assert.Contains(t, result, "●")
}

func TestPrintDeletedCommentSkipped(t *testing.T) {
	t.Parallel()

	comments := &item.Item{
		Title: "Test", User: "poster", Points: 10,
		Comments: []*item.Item{
			{ID: 1, User: "deleted_user", Content: "[deleted]", Level: 0, TimeAgo: "1h ago", Time: 50},
		},
	}
	result := tree.Print(comments, getConfig(), 120, 100)

	assert.NotContains(t, result, "deleted_user")
}

func TestPrintDeletedCommentWithReplies(t *testing.T) {
	t.Parallel()

	comments := &item.Item{
		Title: "Test", User: "poster", Points: 10,
		Comments: []*item.Item{
			{
				ID: 1, User: "deleted_user", Content: "[deleted]", Level: 0, TimeAgo: "1h ago", Time: 50,
				Comments: []*item.Item{
					{ID: 2, User: "carol", Content: "Reply to deleted", Level: 1, TimeAgo: "30m ago", Time: 60},
				},
			},
		},
	}
	result := tree.Print(comments, getConfig(), 120, 100)

	assert.Contains(t, result, "[deleted]")
}

func TestPrintExpandButton(t *testing.T) {
	t.Parallel()

	comments := &item.Item{
		Title: "Test", User: "poster", Points: 10,
		Comments: []*item.Item{
			{
				ID: 1, User: "bob", Content: "Has replies", Level: 0, TimeAgo: "2h ago", Time: 50,
				Comments: []*item.Item{
					{ID: 2, User: "carol", Content: "Reply", Level: 1, TimeAgo: "1h ago", Time: 60},
				},
			},
		},
	}
	result := tree.Print(comments, getConfig(), 120, 100)

	assert.Contains(t, result, "▶")
	assert.Contains(t, result, "1 reply")
}
