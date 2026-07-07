package article

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func paragraph(text string) block {
	return block{kind: blockParagraph, spans: []span{{text: text}}}
}

func TestApplySiteRules_StopAtHeading(t *testing.T) {
	t.Parallel()

	blocks := []block{
		paragraph("keep me"),
		{kind: blockHeading, level: 1, text: "References"},
		paragraph("drop me"),
	}

	out := applySiteRules(blocks, "en.wikipedia.org")

	require.Len(t, out, 1)
	assert.Equal(t, "keep me", out[0].plainText())
}

func TestApplySiteRules_DropBlocks(t *testing.T) {
	t.Parallel()

	blocks := []block{
		paragraph("From Wikipedia, the free encyclopedia"),
		paragraph("actual content"),
		paragraph(`Graham, Paul. "Hacker News Guidelines". Archived from the original on 2020-09-16.`),
	}

	out := applySiteRules(blocks, "en.wikipedia.org")

	require.Len(t, out, 1)
	assert.Equal(t, "actual content", out[0].plainText())
}

func TestApplySiteRules_DropInline(t *testing.T) {
	t.Parallel()

	blocks := []block{
		paragraph("Verified fact.[1] Another claim.[12]"),
		{kind: blockHeading, level: 1, text: "History[edit]"},
		{kind: blockList, items: []listItem{{spans: []span{{text: "item[3]"}}}}},
	}

	out := applySiteRules(blocks, "en.wikipedia.org")

	require.Len(t, out, 3)
	assert.Equal(t, "Verified fact. Another claim.", out[0].plainText())
	assert.Equal(t, "History", out[1].plainText())
	assert.Equal(t, "item", out[2].plainText())
}

func TestApplySiteRules_StopAtBlockContaining(t *testing.T) {
	t.Parallel()

	blocks := []block{
		paragraph("the story"),
		paragraph("This article appeared in the print edition"),
		paragraph("trailing junk"),
	}

	out := applySiteRules(blocks, "www.economist.com")

	require.Len(t, out, 1)
	assert.Equal(t, "the story", out[0].plainText())
}

func TestApplySiteRules_UnknownHostUntouched(t *testing.T) {
	t.Parallel()

	blocks := []block{paragraph("content.[1]")}

	out := applySiteRules(blocks, "example.com")

	require.Len(t, out, 1)
	assert.Equal(t, "content.[1]", out[0].plainText())
}

func TestRulesForHost_MatchesSubdomainsOnly(t *testing.T) {
	t.Parallel()

	_, found := rulesForHost("en.wikipedia.org")
	assert.True(t, found)

	_, found = rulesForHost("wikipedia.org")
	assert.True(t, found)

	_, found = rulesForHost("notwikipedia.org")
	assert.False(t, found)
}

func TestSiteRules_Merge(t *testing.T) {
	t.Parallel()

	shared := siteRules{
		domains:             []string{"example.com"},
		dropBlockContaining: []string{"shared junk"},
	}
	specific := siteRules{
		domains:       []string{"example.com"},
		stopAtHeading: []string{"Related Stories"},
		dropInline:    []*regexp.Regexp{reWikipediaRef},
	}

	merged := shared.merge(specific)

	assert.Equal(t, []string{"shared junk"}, merged.dropBlockContaining)
	assert.Equal(t, []string{"Related Stories"}, merged.stopAtHeading)
	assert.Len(t, merged.dropInline, 1)
}

func TestApplySiteRules_StopAtBlockEquals(t *testing.T) {
	t.Parallel()

	blocks := []block{
		paragraph("the story"),
		paragraph("You may also be interested in:"),
		paragraph("trailing links"),
	}

	out := applySiteRules(blocks, "www.bbc.co.uk")

	require.Len(t, out, 1)
	assert.Equal(t, "the story", out[0].plainText())
}

func TestApplySiteRules_DropBlockMatching(t *testing.T) {
	t.Parallel()

	blocks := []block{
		paragraph("real content"),
		paragraph("84 Comments"),
		paragraph("Comments were 84 times better here"),
	}

	out := applySiteRules(blocks, "arstechnica.com")

	require.Len(t, out, 2)
	assert.Equal(t, "real content", out[0].plainText())
	assert.Equal(t, "Comments were 84 times better here", out[1].plainText())
}

func TestDropInline_LeavesOtherBlocksAlone(t *testing.T) {
	t.Parallel()

	pattern := []*regexp.Regexp{regexp.MustCompile(`x`)}
	b := dropInline(block{kind: blockCode, text: "xyz"}, pattern)

	assert.Equal(t, "yz", b.text)
}
