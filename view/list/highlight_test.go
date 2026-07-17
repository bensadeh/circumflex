package list

import (
	"testing"

	"github.com/bensadeh/circumflex/style"

	"github.com/stretchr/testify/assert"
)

func TestQuerySpans_SubstringsMatchLikeInPageSearch(t *testing.T) {
	spans := querySpans("Testing the tester's tests", "test", false)

	assert.Equal(t, []style.SearchSpan{
		{StartCell: 0, EndCell: 4},
		{StartCell: 12, EndCell: 16},
		{StartCell: 21, EndCell: 25},
	}, spans, "a lowercase query matches caselessly inside every variant")
}

func TestQuerySpans_SmartCase(t *testing.T) {
	spans := querySpans("Rust and rust", "Rust", false)

	assert.Equal(t, []style.SearchSpan{{StartCell: 0, EndCell: 4}}, spans,
		"an uppercase letter in the query makes the match exact")
}

func TestQuerySpans_EachQueryWordMatches(t *testing.T) {
	spans := querySpans("zig before ada", "ada zig", false)

	assert.Equal(t, []style.SearchSpan{
		{StartCell: 0, EndCell: 3},
		{StartCell: 11, EndCell: 14},
	}, spans, "spans come back in line order regardless of word order")
}

func TestQuerySpans_CurrentTierOnSelectedRow(t *testing.T) {
	spans := querySpans("a zig title", "zig", true)

	assert.Equal(t, []style.SearchSpan{{StartCell: 2, EndCell: 5, Current: true}}, spans)
}

func TestQuerySpans_OverlappingWordsKeepFirstClaim(t *testing.T) {
	spans := querySpans("testing", "test sting", false)

	assert.Equal(t, []style.SearchSpan{{StartCell: 0, EndCell: 4}}, spans,
		"'sting' overlaps the cells 'test' already claimed")
}

func TestQuerySpans_NoMatches(t *testing.T) {
	assert.Empty(t, querySpans("nothing here", "absent", false))
	assert.Empty(t, querySpans("nothing here", "", false))
	assert.Empty(t, querySpans("", "word", false))
}

func TestQuerySpans_WideRunesUseCellOffsets(t *testing.T) {
	spans := querySpans("ＡＢＣ title", "title", false)

	assert.Equal(t, []style.SearchSpan{{StartCell: 7, EndCell: 12}}, spans,
		"three double-width runes and a space put the word at cell 7")
}
