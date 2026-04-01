package article

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToMarkdownBlocks_Text(t *testing.T) {
	blocks := convertToMarkdownBlocks("Hello world")

	require.Len(t, blocks, 1)
	assert.Equal(t, blockText, blocks[0].Kind)
	assert.Equal(t, "Hello world", blocks[0].Text)
}

func TestConvertToMarkdownBlocks_MultipleTextParagraphs(t *testing.T) {
	input := "First paragraph\n\nSecond paragraph"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 2)
	assert.Equal(t, blockText, blocks[0].Kind)
	assert.Equal(t, "First paragraph", blocks[0].Text)
	assert.Equal(t, blockText, blocks[1].Kind)
	assert.Equal(t, "Second paragraph", blocks[1].Text)
}

func TestConvertToMarkdownBlocks_TextJoinsLines(t *testing.T) {
	input := "Line one\ncontinued"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockText, blocks[0].Kind)
	assert.Equal(t, "Line one continued", blocks[0].Text)
}

func TestConvertToMarkdownBlocks_Headers(t *testing.T) {
	tests := []struct {
		input    string
		kind     int
		wantText string
	}{
		{"# Title", blockH1, "# Title"},
		{"## Subtitle", blockH2, "## Subtitle"},
		{"### Section", blockH3, "### Section"},
		{"#### Sub-section", blockH4, "#### Sub-section"},
		{"##### Minor", blockH5, "##### Minor"},
		{"###### Tiny", blockH6, "###### Tiny"},
	}

	for _, tt := range tests {
		blocks := convertToMarkdownBlocks(tt.input)
		require.Len(t, blocks, 1, "input: %s", tt.input)
		assert.Equal(t, tt.kind, blocks[0].Kind, "input: %s", tt.input)
		assert.Equal(t, tt.wantText, blocks[0].Text, "input: %s", tt.input)
	}
}

func TestConvertToMarkdownBlocks_Code(t *testing.T) {
	input := "```\nfunc main() {}\n```"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockCode, blocks[0].Kind)
	assert.Contains(t, blocks[0].Text, "func main() {}")
}

func TestConvertToMarkdownBlocks_Quote(t *testing.T) {
	input := "> This is a quote"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockQuote, blocks[0].Kind)
	assert.Equal(t, "This is a quote", blocks[0].Text)
}

func TestConvertToMarkdownBlocks_List(t *testing.T) {
	input := "- item one\n- item two"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockList, blocks[0].Kind)
	assert.Contains(t, blocks[0].Text, "item one")
	assert.Contains(t, blocks[0].Text, "item two")
}

func TestConvertToMarkdownBlocks_NumberedList(t *testing.T) {
	input := "1. first\n2. second"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockList, blocks[0].Kind)
}

func TestConvertToMarkdownBlocks_Table(t *testing.T) {
	input := "| A | B |\n| - | - |\n| 1 | 2 |"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockTable, blocks[0].Kind)
}

func TestConvertToMarkdownBlocks_Divider(t *testing.T) {
	input := "* * *"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockDivider, blocks[0].Kind)
}

func TestConvertToMarkdownBlocks_Image(t *testing.T) {
	input := "![alt text](https://example.com/img.png)"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockImage, blocks[0].Kind)
}

func TestConvertToMarkdownBlocks_EmptyInput(t *testing.T) {
	blocks := convertToMarkdownBlocks("")
	assert.Empty(t, blocks)
}

func TestConvertToMarkdownBlocks_ReplacesEnDashAndEmDash(t *testing.T) {
	input := "word–word—word"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 1)
	assert.NotContains(t, blocks[0].Text, "–")
	assert.NotContains(t, blocks[0].Text, "—")
	assert.Contains(t, blocks[0].Text, "-")
}

func TestConvertToMarkdownBlocks_MultiBlockQuote(t *testing.T) {
	input := "> line one\n> line two"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 1)
	assert.Equal(t, blockQuote, blocks[0].Kind)
	assert.Contains(t, blocks[0].Text, "line one")
	assert.Contains(t, blocks[0].Text, "line two")
}

func TestConvertToMarkdownBlocks_Mixed(t *testing.T) {
	input := "# Header\n\nSome text.\n\n```\ncode\n```\n\n> quote"
	blocks := convertToMarkdownBlocks(input)

	require.Len(t, blocks, 4)
	assert.Equal(t, blockH1, blocks[0].Kind)
	assert.Equal(t, blockText, blocks[1].Kind)
	assert.Equal(t, blockCode, blocks[2].Kind)
	assert.Equal(t, blockQuote, blocks[3].Kind)
}

func TestIsListItem(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"- item", true},
		{"1. item", true},
		{"  - nested", true},
		{"  10. numbered", true},
		{"not a list", false},
		{"", false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, isListItem(tt.input), "input: %q", tt.input)
	}
}

func TestConvertToTerminalFormat_BasicText(t *testing.T) {
	blocks := []*block{
		{Kind: blockText, Text: "Hello world"},
	}

	result := convertToTerminalFormat(blocks, 80, "│")
	assert.Contains(t, result, "Hello world")
}

func TestConvertToTerminalFormat_MultiplBlocks(t *testing.T) {
	blocks := []*block{
		{Kind: blockText, Text: "First paragraph"},
		{Kind: blockText, Text: "Second paragraph"},
	}

	result := convertToTerminalFormat(blocks, 80, "│")
	assert.Contains(t, result, "First paragraph")
	assert.Contains(t, result, "Second paragraph")
	// Blocks are separated by double newlines
	assert.Contains(t, result, "\n\n")
}

func TestConvertToTerminalFormat_Code(t *testing.T) {
	blocks := []*block{
		{Kind: blockCode, Text: "\nfmt.Println()\n"},
	}

	result := convertToTerminalFormat(blocks, 80, "│")
	assert.Contains(t, result, "fmt.Println()")
}

func TestConvertToTerminalFormat_Divider(t *testing.T) {
	blocks := []*block{
		{Kind: blockDivider, Text: "* * *"},
	}

	result := convertToTerminalFormat(blocks, 80, "│")
	assert.Contains(t, result, strings.Repeat("-", 72))
}

func TestUnescapeCharacters(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`\|`, "|"},
		{`\-`, "-"},
		{`\_`, "_"},
		{`\*`, "*"},
		{`\\`, `\`},
		{`\#`, "#"},
		{`\.`, "."},
		{`\>`, ">"},
		{`\<`, "<"},
		{"\\`", "`"},
		{"...", "…"},
		{`\(`, "("},
		{`\)`, ")"},
		{`\[`, "["},
		{`\]`, "]"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, unescapeCharacters(tt.input), "input: %q", tt.input)
	}
}

func TestRemoveImageReference(t *testing.T) {
	input := "Text ![alt](http://img.png) more"
	result := removeImageReference(input)
	assert.Equal(t, "Text alt more", result)
}

func TestRemoveHrefs(t *testing.T) {
	input := `Click <a href="http://example.com">here</a>`
	result := removeHrefs(input)
	assert.Equal(t, "Click here", result)
}

func TestIt_ReplacesItalicMarkers(t *testing.T) {
	input := italicStart + "italic text" + italicStop
	result := it(input)
	assert.Contains(t, result, "\u001B[3m")
	assert.Contains(t, result, "\u001B[23m")
	assert.Contains(t, result, "italic text")
}

func TestPreFormatHeader(t *testing.T) {
	input := "## My Header"
	result := preFormatHeader(input)
	assert.Equal(t, "My Header", result)
}

func TestTrimLeadingZero(t *testing.T) {
	input := indentLevel2 + "01. item"
	result := trimLeadingZero(input)
	assert.Equal(t, indentLevel2+" 1. item", result)
}
