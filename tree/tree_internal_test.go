package tree

import (
	"strings"
	"testing"

	"clx/constants/unicode"
	"clx/item"

	"github.com/stretchr/testify/assert"
)

func TestGetIndentString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "", getIndentString(0))
	assert.Equal(t, "", getIndentString(1))
	assert.Equal(t, " ", getIndentString(2))
	assert.Equal(t, "  ", getIndentString(3))
}

func TestGetSeparator(t *testing.T) {
	t.Parallel()

	t.Run("first comment returns empty", func(t *testing.T) {
		assert.Equal(t, "", getSeparator(0, 10, 42, 42))
	})

	t.Run("level 0 non-first has underline separator", func(t *testing.T) {
		result := getSeparator(0, 5, 99, 42)
		assert.Contains(t, result, "▁")
		assert.True(t, strings.HasSuffix(result, "\n\n"))
	})

	t.Run("level > 0 returns newline only", func(t *testing.T) {
		assert.Equal(t, "\n", getSeparator(1, 10, 99, 42))
	})
}

func TestGetAuthorOldComment(t *testing.T) {
	t.Parallel()

	result := getAuthor("alice", 100, 50)

	assert.Contains(t, result, "alice")
	assert.NotContains(t, result, "●")
}

func TestGetAuthorNewComment(t *testing.T) {
	t.Parallel()

	result := getAuthor("alice", 50, 100)

	assert.Contains(t, result, "alice")
	assert.Contains(t, result, "●")
}

func TestGetAuthorLabelMod(t *testing.T) {
	t.Parallel()

	result := getAuthorLabel("dang", "someone", "", false)

	assert.Contains(t, result, "mod")
}

func TestGetAuthorLabelOP(t *testing.T) {
	t.Parallel()

	result := getAuthorLabel("alice", "alice", "", false)

	assert.Contains(t, result, "OP")
}

func TestGetAuthorLabelPP(t *testing.T) {
	t.Parallel()

	result := getAuthorLabel("bob", "alice", "bob", false)

	assert.Contains(t, result, "PP")
}

func TestGetAuthorLabelRegular(t *testing.T) {
	t.Parallel()

	result := getAuthorLabel("charlie", "alice", "bob", false)

	assert.Equal(t, "", result)
}

func TestIsMod(t *testing.T) {
	t.Parallel()

	assert.True(t, isMod("dang"))
	assert.True(t, isMod("tomhow"))
	assert.False(t, isMod("other"))
}

func TestGetReplyCount(t *testing.T) {
	t.Parallel()

	root := &item.Item{
		Comments: []*item.Item{
			{
				Comments: []*item.Item{
					{Comments: nil},
				},
			},
			{Comments: nil},
		},
	}

	assert.Equal(t, 3, getReplyCount(root))
}

func TestGetNewCommentsCount(t *testing.T) {
	t.Parallel()

	root := &item.Item{
		Comments: []*item.Item{
			{Time: 200, Comments: []*item.Item{
				{Time: 300, Comments: nil},
			}},
			{Time: 50, Comments: nil},
		},
	}

	// lastVisited=100: Time 200 and 300 are new, 50 is old
	assert.Equal(t, 2, getNewCommentsCount(root, 100))
}

func TestGetButton(t *testing.T) {
	t.Parallel()

	t.Run("level 0 with replies", func(t *testing.T) {
		result := getButton(0, 3, 80)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "3 replies")
	})

	t.Run("level 0 no replies", func(t *testing.T) {
		assert.Empty(t, getButton(0, 0, 80))
	})

	t.Run("level > 0", func(t *testing.T) {
		assert.Empty(t, getButton(1, 5, 80))
	})

	t.Run("singular reply", func(t *testing.T) {
		result := getButton(0, 1, 80)
		assert.Contains(t, result, "1 reply")
		assert.NotContains(t, result, "replies")
	})
}

func TestAddFilterTag(t *testing.T) {
	t.Parallel()

	t.Run("level 0 unchanged", func(t *testing.T) {
		input := "line1\nline2\n"
		assert.Equal(t, input, addFilterTag(0, input))
	})

	t.Run("level > 0 adds invisible chars", func(t *testing.T) {
		input := "line1\nline2\n"
		result := addFilterTag(1, input)
		assert.Contains(t, result, unicode.InvisibleCharacterForExpansion)
	})
}

func TestGetZeroWidthSpace(t *testing.T) {
	t.Parallel()

	assert.Equal(t, unicode.InvisibleCharacterForTopLevelComments, getZeroWidthSpace(true))
	assert.Equal(t, "", getZeroWidthSpace(false))
}

func TestGetFirstCommentID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 0, getFirstCommentID(nil))
	assert.Equal(t, 0, getFirstCommentID([]*item.Item{}))
	assert.Equal(t, 42, getFirstCommentID([]*item.Item{{ID: 42}, {ID: 99}}))
}
