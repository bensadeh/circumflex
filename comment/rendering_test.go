package comment

import (
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/ansi"

	"github.com/stretchr/testify/assert"
)

func TestIndentString(t *testing.T) {
	t.Parallel()

	assert.Empty(t, IndentString(0))
	assert.Empty(t, IndentString(1))
	assert.Equal(t, " ", IndentString(2))
	assert.Equal(t, "  ", IndentString(3))
}

func TestSeparator(t *testing.T) {
	t.Parallel()

	t.Run("first comment returns empty", func(t *testing.T) {
		assert.Empty(t, Separator(0, 10, 42, 42))
	})

	t.Run("level 0 non-first has underline separator", func(t *testing.T) {
		result := Separator(0, 5, 99, 42)
		assert.Contains(t, result, "\u2581")
		assert.True(t, strings.HasSuffix(result, "\n\n"))
	})

	t.Run("level > 0 returns newline only", func(t *testing.T) {
		assert.Equal(t, "\n", Separator(1, 10, 99, 42))
	})
}

func TestAuthorOldComment(t *testing.T) {
	t.Parallel()

	result := Author("alice", 100, 50, false)

	assert.Contains(t, result, "alice")
	assert.NotContains(t, result, "\u25cf")
}

func TestAuthorNewComment(t *testing.T) {
	t.Parallel()

	result := Author("alice", 50, 100, false)

	assert.Contains(t, result, "alice")
	assert.Contains(t, result, "\u25cf")
}

func TestAuthorLabelMod(t *testing.T) {
	t.Parallel()

	result := AuthorLabel("dang", "someone", "", false)

	assert.Contains(t, result, "mod")
}

func TestAuthorLabelOP(t *testing.T) {
	t.Parallel()

	result := AuthorLabel("alice", "alice", "", false)

	assert.Contains(t, result, "OP")
}

func TestAuthorLabelGP(t *testing.T) {
	t.Parallel()

	result := AuthorLabel("bob", "alice", "bob", false)

	assert.Contains(t, result, "GP")
}

func TestAuthorLabelRegular(t *testing.T) {
	t.Parallel()

	result := AuthorLabel("charlie", "alice", "bob", false)

	assert.Empty(t, result)
}

func TestIsMod(t *testing.T) {
	t.Parallel()

	assert.True(t, IsMod("dang"))
	assert.True(t, IsMod("tomhow"))
	assert.False(t, IsMod("other"))
}

func TestNewCommentsCount(t *testing.T) {
	t.Parallel()

	thread := &Thread{
		Comments: []*Comment{
			{Time: 200, Children: []*Comment{
				{Time: 300},
			}},
			{Time: 50},
		},
	}

	// lastVisited=100: Time 200 and 300 are new, 50 is old
	assert.Equal(t, 2, NewCommentsCount(thread, 100))
}

func TestFirstCommentID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, FirstCommentID(nil))
	assert.Equal(t, 0, FirstCommentID([]*Comment{}))
	assert.Equal(t, 42, FirstCommentID([]*Comment{{ID: 42}, {ID: 99}}))
}

func TestRepliesIndicator(t *testing.T) {
	t.Parallel()

	t.Run("singular reply", func(t *testing.T) {
		result := RepliesIndicator(1, 0, 80, true)
		assert.Contains(t, result, "1 reply")
		assert.NotContains(t, result, "replies")
	})

	t.Run("plural replies", func(t *testing.T) {
		result := RepliesIndicator(3, 0, 80, false)
		assert.Contains(t, result, "3 replies")
	})

	t.Run("collapsed and expanded have same visible length", func(t *testing.T) {
		collapsed := ansi.Strip(RepliesIndicator(5, 0, 80, true))
		expanded := ansi.Strip(RepliesIndicator(5, 0, 80, false))
		assert.Len(t, expanded, len(collapsed))
	})
}
