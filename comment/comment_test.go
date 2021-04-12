package comment_test

import (
	"clx/comment"
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
	actual := comment.ToString(*comments, 4, 80, 200, false)

	assert.Equal(t, string(expected), actual)
}

func TestRootComment(t *testing.T) {
	t.Parallel()

	commentJSON, _ := ioutil.ReadFile("test/root_comment.json")
	expected, _ := ioutil.ReadFile("test/root_comment_expected.txt")

	comments := unmarshal(commentJSON)
	actual := comment.ToString(*comments, 4, 80, 200, false)

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

	expected := "Not a code Block: \n\n\x1b[2m  CODE BLOCK CODE BLOCK \nCODE BLOCK CODE BLOCK\x1b[0m"

	actual, _ := comment.ParseComment(input)

	assert.Equal(t, expected, actual)
}
