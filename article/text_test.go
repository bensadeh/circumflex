package article

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTextBlocks_PreservesLineStructure(t *testing.T) {
	t.Parallel()

	input := "OpenSSH 10.4 release notes\r\n\r\n" +
		"Changes since 10.3:\r\n" +
		" * ssh(1): fix a bug\r\n" +
		"   with continuation\r\n\r\n\r\n" +
		"Checksums:   \r\n\tSHA256 abc\r\n"

	blocks := parseTextBlocks(input)

	require.Len(t, blocks, 3)

	for _, b := range blocks {
		assert.Equal(t, blockVerbatim, b.kind)
	}

	assert.Equal(t, "OpenSSH 10.4 release notes", blocks[0].text)
	assert.Equal(t, "Changes since 10.3:\n * ssh(1): fix a bug\n   with continuation", blocks[1].text)
	assert.Equal(t, "Checksums:\n        SHA256 abc", blocks[2].text, "tabs expand, trailing spaces trimmed")
}

func TestRenderBlocks_VerbatimKeepsIndentation(t *testing.T) {
	t.Parallel()

	blocks := []block{{kind: blockVerbatim, text: " * item one\n   continuation"}}

	assert.Equal(t, " * item one\n   continuation", renderBlocks(blocks, 72))
}

func TestRenderBlocks_VerbatimWrapsLongLines(t *testing.T) {
	t.Parallel()

	blocks := []block{{kind: blockVerbatim, text: strings.Repeat("word ", 20)}}

	for line := range strings.SplitSeq(renderBlocks(blocks, 30), "\n") {
		assert.LessOrEqual(t, len(line), 30)
	}
}

func TestIsPlainText(t *testing.T) {
	t.Parallel()

	assert.True(t, isPlainText("text/plain", []byte("release notes")))
	assert.True(t, isPlainText("text/plain; charset=utf-8", []byte("notes")))
	assert.False(t, isPlainText("text/html", []byte("plain looking")))
	assert.False(t, isPlainText("text/plain", []byte("<!DOCTYPE html><p>mislabeled</p>")),
		"mislabeled HTML should go through readability")
}
