package article

import (
	nurl "net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullTextURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url  string
		want string
	}{
		{url: "https://arxiv.org/abs/2607.06377", want: "https://arxiv.org/html/2607.06377"},
		{url: "https://arxiv.org/abs/2607.06377v2", want: "https://arxiv.org/html/2607.06377v2"},
		{url: "https://www.arxiv.org/abs/2607.06377", want: "https://arxiv.org/html/2607.06377"},
		{url: "https://arxiv.org/pdf/2607.06377", want: "https://arxiv.org/html/2607.06377"},
		{url: "https://arxiv.org/pdf/2607.06377v1.pdf", want: "https://arxiv.org/html/2607.06377v1"},
		{url: "https://arxiv.org/abs/quant-ph/0410100", want: "https://arxiv.org/html/quant-ph/0410100"},
		{url: "https://arxiv.org/abs/2607.06377?context=math", want: "https://arxiv.org/html/2607.06377"},
		{url: "https://export.arxiv.org/abs/2607.06377", want: "https://arxiv.org/html/2607.06377"},
		{url: "https://arxiv.org/html/2607.06377", want: ""},
		{url: "https://arxiv.org/list/math.HO/recent", want: ""},
		{url: "https://arxiv.org", want: ""},
		{url: "https://example.com/abs/2607.06377", want: ""},
		{url: "https://notarxiv.org/abs/2607.06377", want: ""},
	}

	for _, tt := range tests {
		parsed, err := nurl.Parse(tt.url)
		require.NoError(t, err)

		assert.Equal(t, tt.want, fullTextURL(parsed), tt.url)
	}
}

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

func TestApplySiteRules_StopAtHeadingIgnoresTrailingWhitespace(t *testing.T) {
	t.Parallel()

	blocks := []block{
		paragraph("body"),
		{kind: blockHeading, level: 1, text: "References [1]"},
		paragraph("citation dump"),
	}

	out := applySiteRules(blocks, "en.wikipedia.org")

	require.Len(t, out, 1, "heading with trailing space after [1] removal must still stop the article")
	assert.Equal(t, "body", out[0].plainText())
}

func TestDropInline_CleansTableCells(t *testing.T) {
	t.Parallel()

	blocks := []block{
		{kind: blockTable, rows: [][]string{{"Year", "Value"}, {"2007[1]", "Online[2]"}}},
	}

	out := applySiteRules(blocks, "en.wikipedia.org")

	require.Len(t, out, 1)
	assert.Equal(t, [][]string{{"Year", "Value"}, {"2007", "Online"}}, out[0].rows)
}

func TestDropInline_SkipsCodeBlocks(t *testing.T) {
	t.Parallel()

	pattern := []*regexp.Regexp{regexp.MustCompile(`\[\d+\]`)}

	code := dropInline(block{kind: blockCode, text: "arr[1]"}, pattern)
	assert.Equal(t, "arr[1]", code.text, "citation strippers must not rewrite code")

	verbatim := dropInline(block{kind: blockVerbatim, text: "arr[1]"}, pattern)
	assert.Equal(t, "arr[1]", verbatim.text)

	prose := dropInline(block{kind: blockParagraph, spans: []span{{text: "fact[1]"}}}, pattern)
	assert.Equal(t, "fact", prose.plainText(), "prose is still cleaned")
}
