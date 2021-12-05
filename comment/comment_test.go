package comment_test

import (
	"clx/comment"
	"clx/core"
	"clx/item"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullParse(t *testing.T) {
	t.Parallel()

	commentJSON, _ := ioutil.ReadFile("test/comments.json")
	expected, _ := ioutil.ReadFile("test/expected.txt")

	comments := unmarshal(commentJSON)
	actual := comment.ToString(comments, getConfig(), 85)

	assert.Equal(t, string(expected), actual)
}

func unmarshal(data []byte) *item.Item {
	root := new(item.Item)
	_ = json.Unmarshal(data, &root)

	return root
}

func getConfig() *core.Config {
	return &core.Config{
		CommentWidth:       80,
		HighlightHeadlines: true,
		RelativeNumbering:  false,
		HideYCJobs:         false,
		HighlightComments:  true,
		EmojiSmileys:       true,
		MarkAsRead:         false,
		IndentationSymbol:  "▎",
	}
}
