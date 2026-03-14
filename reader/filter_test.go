package reader

import (
	"clx/constants"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter_SkipParContains(t *testing.T) {
	t.Parallel()

	rs := ruleSet{}
	rs.skipParContains("junk")

	input := "First paragraph\n\nThis has junk in it\n\nThird paragraph"
	result := rs.filter(input)

	assert.Contains(t, result, "First paragraph")
	assert.NotContains(t, result, "junk")
	assert.Contains(t, result, "Third paragraph")
}

func TestFilter_SkipLineEquals(t *testing.T) {
	t.Parallel()

	rs := ruleSet{}
	rs.skipLineEquals("Credit")

	input := "First line\nContent\nCredit\nMore content\nLast line"
	result := rs.filter(input)

	assert.Contains(t, result, "First line")
	assert.Contains(t, result, "Content")
	assert.NotContains(t, result, "Credit")
	assert.Contains(t, result, "More content")
}

func TestFilter_EndBeforeLineEquals(t *testing.T) {
	t.Parallel()

	// endBefore checks the NEXT line — when it matches, the current line
	// and everything after are dropped.
	rs := ruleSet{}
	rs.endBeforeLineEquals("References")

	input := "Content one\n\nContent two\n\nReferences\n\nRef list"
	result := rs.filter(input)

	assert.Contains(t, result, "Content one")
	assert.Contains(t, result, "Content two")
	assert.NotContains(t, result, "References")
	assert.NotContains(t, result, "Ref list")
}

func TestFilter_EndBeforeLineContains(t *testing.T) {
	t.Parallel()

	rs := ruleSet{}
	rs.endBeforeLineContains("appeared in the")

	input := "Content one\n\nContent two\n\nThis article appeared in the print edition\n\nMore"
	result := rs.filter(input)

	assert.Contains(t, result, "Content one")
	assert.Contains(t, result, "Content two")
	assert.NotContains(t, result, "appeared in the")
}

func TestFilter_SkipsSingleCharLines(t *testing.T) {
	t.Parallel()

	rs := ruleSet{}
	input := "First line\n \nThird line"
	result := rs.filter(input)

	assert.Contains(t, result, "First line")
	assert.Contains(t, result, "Third line")
}

func TestFilter_NoRules_PassesThrough(t *testing.T) {
	t.Parallel()

	rs := ruleSet{}
	input := "Line one\nLine two\nLine three"
	result := rs.filter(input)

	assert.Contains(t, result, "Line one")
	assert.Contains(t, result, "Line two")
	assert.Contains(t, result, "Line three")
}

func TestFilter_CollapsesExcessiveNewlines(t *testing.T) {
	t.Parallel()

	rs := ruleSet{}
	rs.skipParContains("remove me")

	input := "First\n\nremove me\n\nSecond"
	result := rs.filter(input)

	assert.NotContains(t, result, "\n\n\n")
}

func TestFilter_MultipleRulesCombined(t *testing.T) {
	t.Parallel()

	rs := ruleSet{}
	rs.skipParContains("Subscribe")
	rs.skipLineEquals("Image")

	input := "Title\n\nSubscribe to our newsletter\n\nImage\nReal content\nLast line"
	result := rs.filter(input)

	assert.Contains(t, result, "Title")
	assert.NotContains(t, result, "Subscribe")
	assert.NotContains(t, result, "Image")
	assert.Contains(t, result, "Real content")
}

func TestFilterSite_UnknownSite_PassesThrough(t *testing.T) {
	t.Parallel()

	input := "Article content here"
	result := filterSite(input, "https://example.com/article")

	assert.Equal(t, input, result)
}

func TestFilterSite_Wikipedia_StripsBracketedEdit(t *testing.T) {
	t.Parallel()

	input := "History [edit]\n\nSome content"
	result := filterSite(input, "https://en.wikipedia.org/wiki/Test")

	assert.NotContains(t, result, "[edit]")
	assert.Contains(t, result, "History")
}

func TestFilterSite_Wikipedia_RemovesNumberedReferences(t *testing.T) {
	t.Parallel()

	input := "Some fact[1] and another[23]\n\nMore content"
	result := filterSite(input, "https://en.wikipedia.org/wiki/Test")

	assert.NotContains(t, result, "[1]")
	assert.NotContains(t, result, "[23]")
	assert.Contains(t, result, "Some fact")
}

func TestFilterSite_Wikipedia_EndsBeforeReferences(t *testing.T) {
	t.Parallel()

	input := "Intro\n\nArticle content\n\n" + constants.Block + " References\n\nRef list"
	result := filterSite(input, "https://en.wikipedia.org/wiki/Test")

	assert.Contains(t, result, "Article content")
	assert.NotContains(t, result, "Ref list")
}

func TestFilterSite_Wikipedia_EndsBeforeSeeAlso(t *testing.T) {
	t.Parallel()

	input := "Intro\n\nArticle content\n\n" + constants.Block + " See also\n\nMore links"
	result := filterSite(input, "https://en.wikipedia.org/wiki/Test")

	assert.Contains(t, result, "Article content")
	assert.NotContains(t, result, "More links")
}

func TestFilterSite_NYTimes_SkipsCreditParagraphs(t *testing.T) {
	t.Parallel()

	input := "News content\n\nCredit… John Doe for NYT\n\nMore news"
	result := filterSite(input, "https://www.nytimes.com/article")

	assert.Contains(t, result, "News content")
	assert.NotContains(t, result, "Credit…")
	assert.Contains(t, result, "More news")
}

func TestFilterSite_NYTimes_SkipsCreditAndImageLines(t *testing.T) {
	t.Parallel()

	input := "First line\nCredit\nImage\nReal content\nLast line"
	result := filterSite(input, "https://www.nytimes.com/article")

	assert.Contains(t, result, "Real content")
	assert.NotContains(t, result, "\nCredit\n")
	assert.NotContains(t, result, "\nImage\n")
}

func TestFilterSite_Economist_SkipsListenPrompts(t *testing.T) {
	t.Parallel()

	input := "Article\n\nListen to this story on Spotify\n\nReal content"
	result := filterSite(input, "https://www.economist.com/article")

	assert.Contains(t, result, "Article")
	assert.NotContains(t, result, "Listen to this story")
	assert.Contains(t, result, "Real content")
}

func TestFilterSite_Economist_EndsBeforeArticleAppearedIn(t *testing.T) {
	t.Parallel()

	input := "Intro\n\nArticle content\n\nThis article appeared in the print edition\n\nMore"
	result := filterSite(input, "https://www.economist.com/article")

	assert.Contains(t, result, "Article content")
	assert.NotContains(t, result, "appeared in the print edition")
}

func TestFilterSite_ArsTechnica_SkipsEnlarge(t *testing.T) {
	t.Parallel()

	input := "Article\n\nEnlarge/ Some image caption\n\nReal content"
	result := filterSite(input, "https://arstechnica.com/article")

	assert.NotContains(t, result, "Enlarge/")
	assert.Contains(t, result, "Real content")
}

func TestFilterSite_Wired_EndsBeforeMoreStories(t *testing.T) {
	t.Parallel()

	input := "Intro\n\nArticle content\n\nMore Great WIRED Stories\n\nLink list"
	result := filterSite(input, "https://www.wired.com/article")

	assert.Contains(t, result, "Article content")
	assert.NotContains(t, result, "More Great WIRED Stories")
}

func TestFilterSite_Guardian_SkipsPhotographCredits(t *testing.T) {
	t.Parallel()

	input := "Article\n\nPhotograph: Getty Images\n\nReal content"
	result := filterSite(input, "https://www.theguardian.com/article")

	assert.NotContains(t, result, "Photograph:")
	assert.Contains(t, result, "Real content")
}
