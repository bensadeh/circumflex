package article

import (
	nurl "net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pkg.go.dev-shaped declarations: comments in spans classed "comment", field
// lines in spans whose ids carry chrome words (RelatedInformation.Pos), every
// type name a cross-package link. Readability would delete the spans as
// unlikely candidates and the second declaration's wrapper as a link-heavy
// div; the pre text must come through whole.
func TestExtractReadable_PreMarkupSurvives(t *testing.T) {
	t.Parallel()

	// The prose must clear readability's 500-char threshold on its own:
	// shorter pages get re-parsed with the chrome heuristics disabled, and
	// the damage this test pins never happens.
	filler := `<p>Enough readable content for the extractor to accept this page as an article, ` +
		`repeated to pass its length heuristics without any help from the code blocks below.</p>`

	page := `<html><head><title>T</title></head><body><article>` +
		filler + filler + filler + filler + filler +
		`<div><pre>type RelatedInformation struct {
<span id="RelatedInformation.Pos" data-kind="field">	Pos <a href="/go/token">token</a>.<a href="/go/token#Pos">Pos</a> <span class="comment">// start position</span>
</span>}</pre></div>` +
		`<div><pre>func (pass *<a href="/tools/go/analysis#Pass">Pass</a>) Reportf(pos <a href="/go/token">token</a>.<a href="/go/token#Pos">Pos</a>, format <a href="/builtin#string">string</a>, args ...<a href="/builtin#any">any</a>)</pre></div>` +
		`</article></body></html>`

	u, err := nurl.Parse("https://pkg.go.dev/example")
	require.NoError(t, err)

	node, _, err := extractReadable([]byte(page), u)
	require.NoError(t, err)

	var codes []string

	for _, b := range parseBlocks(node) {
		if b.kind == blockCode {
			codes = append(codes, b.text)
		}
	}

	require.Len(t, codes, 2, "both declarations must survive readability")
	assert.Contains(t, codes[0], "Pos token.Pos // start position")
	assert.Contains(t, codes[1], "func (pass *Pass) Reportf(pos token.Pos, format string, args ...any)")
}
