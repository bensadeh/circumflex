package comment_test

import (
	"clx/comment"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
	commentJSON, _ := ioutil.ReadFile("test/comments.json")
	expected, _ := ioutil.ReadFile("test/expected.txt")

	comments := unmarshal(commentJSON)
	actual := comment.ToString(*comments, 4, 80, 200, false)

	assert.Equal(t, string(expected), actual)
}

func unmarshal(data []byte) *comment.Comments {
	root := new(comment.Comments)
	_ = json.Unmarshal(data, &root)

	return root
}
