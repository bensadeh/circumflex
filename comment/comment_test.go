package comment_test

import (
	"clx/comment"
	"clx/core"
	"clx/endpoints"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
	t.Parallel()

	commentJSON, _ := ioutil.ReadFile("test/comments.json")
	expected, _ := ioutil.ReadFile("test/expected.txt")

	comments := unmarshal(commentJSON)
	actual := comment.ToString(*comments, getConfig(), 100)

	assert.Equal(t, string(expected), actual)
}

func TestRootComment(t *testing.T) {
	t.Parallel()

	commentJSON, _ := ioutil.ReadFile("test/root_comment.json")
	expected, _ := ioutil.ReadFile("test/root_comment_expected.txt")

	comments := unmarshal(commentJSON)
	actual := comment.ToString(*comments, getConfig(), 100)

	assert.Equal(t, string(expected), actual)
}

func unmarshal(data []byte) *endpoints.Comments {
	root := new(endpoints.Comments)
	_ = json.Unmarshal(data, &root)

	return root
}

func TestParsing(t *testing.T) {
	t.Parallel()

	input := "<p>Not a code Block: " +
		"<p><pre><code>  CODE BLOCK CODE BLOCK \n" +
		"CODE BLOCK CODE BLOCK</code></pre>"

	expected := "Not a code Block:\n\n\x1b[2m80  CODE BLOCK CODE BLOCK\x1b[0m\n\x1b[2m80CODE BLOCK CODE BLOCK\x1b[0m"

	actual := comment.ParseComment(input, getConfig(), 80, 80)

	assert.Equal(t, expected, actual)
}

func getConfig() *core.Config {
	return &core.Config{
		CommentWidth:         80,
		IndentSize:           4,
		HighlightHeadlines:   true,
		PreserveCommentWidth: false,
		RelativeNumbering:    false,
		HideYCJobs:           false,
		AltIndentBlock:       false,
		HighlightComments:    true,
		EmojiSmileys:         true,
		MarkAsRead:           false,
	}
}
