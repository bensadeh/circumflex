package comment

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEffectiveIndentColumns(t *testing.T) {
	t.Parallel()

	t.Run("top level and first nest have zero indent", func(t *testing.T) {
		assert.Equal(t, 0, EffectiveIndentColumns(0, 4, 70, 40))
		assert.Equal(t, 0, EffectiveIndentColumns(1, 4, 70, 40))
	})

	t.Run("scales linearly while under the floor", func(t *testing.T) {
		assert.Equal(t, 4, EffectiveIndentColumns(2, 4, 70, 40))
		assert.Equal(t, 8, EffectiveIndentColumns(3, 4, 70, 40))
	})

	t.Run("plateaus at commentWidth minus minCommentWidth", func(t *testing.T) {
		// headroom = 70 - 40 = 30; (20-1)*4 = 76 desired, capped at 30
		assert.Equal(t, 30, EffectiveIndentColumns(20, 4, 70, 40))
		// Every deeper comment stays at the same plateau.
		assert.Equal(t, 30, EffectiveIndentColumns(50, 4, 70, 40))
	})

	t.Run("commentWidth at or below floor collapses indent to zero", func(t *testing.T) {
		assert.Equal(t, 0, EffectiveIndentColumns(10, 4, 40, 40))
		assert.Equal(t, 0, EffectiveIndentColumns(10, 4, 20, 40))
	})
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

	t.Run("singular reply hidden", func(t *testing.T) {
		result := RepliesIndicator(1, "", true)
		assert.Contains(t, result, "1 reply hidden")
		assert.NotContains(t, result, "replies")
	})

	t.Run("plural replies hidden", func(t *testing.T) {
		result := RepliesIndicator(3, "", true)
		assert.Contains(t, result, "3 replies hidden")
	})

	t.Run("expanded returns empty", func(t *testing.T) {
		result := RepliesIndicator(5, "", false)
		assert.Empty(t, result)
	})

	t.Run("caller-supplied indent is respected", func(t *testing.T) {
		result := RepliesIndicator(2, "    ", true)
		assert.Contains(t, result, "    ")
	})
}
