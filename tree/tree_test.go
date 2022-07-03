package tree_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"clx/item"
	"clx/settings"
	"clx/tree"

	"github.com/stretchr/testify/assert"
)

func TestPring(t *testing.T) {
	t.Parallel()

	commentJSON, _ := ioutil.ReadFile("test/comments.json")
	expected, _ := ioutil.ReadFile("test/expected.txt")

	comments := unmarshal(commentJSON)
	actual := tree.Print(comments, getConfig(), 85, 1643215106)

	assert.Equal(t, string(expected), actual)
}

func unmarshal(data []byte) *item.Item {
	root := new(item.Item)
	_ = json.Unmarshal(data, &root)

	return root
}

func getConfig() *settings.Config {
	return &settings.Config{
		CommentWidth:       80,
		HighlightHeadlines: true,
		RelativeNumbering:  false,
		HighlightComments:  true,
		EmojiSmileys:       true,
		MarkAsRead:         false,
		IndentationSymbol:  "▎",
	}
}
