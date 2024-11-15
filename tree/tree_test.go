package tree_test

import (
	"encoding/json"
	"charm.land/lipgloss/v2"
	"github.com/muesli/termenv"
	"os"
	"testing"

	"clx/item"
	"clx/settings"
	"clx/tree"

	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	t.Parallel()

	lipgloss.SetColorProfile(termenv.ANSI)

	commentJSON, _ := os.ReadFile("test/comments.json")
	expected, _ := os.ReadFile("test/expected.txt")

	comments := unmarshal(commentJSON)
	actual := tree.Print(comments, getConfig(), 120, 1643215106)

	assert.Equal(t, string(expected), actual)
}

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
