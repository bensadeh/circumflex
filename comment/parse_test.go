package comment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func kinds(blocks []Block) []blockKind {
	out := make([]blockKind, len(blocks))
	for i, b := range blocks {
		out[i] = b.kind
	}

	return out
}

func spanTexts(b Block) []string {
	out := make([]string, len(b.spans))
	for i, s := range b.spans {
		out[i] = s.text
	}

	return out
}

func spanFormats(b Block) []spanFormat {
	out := make([]spanFormat, len(b.spans))
	for i, s := range b.spans {
		out[i] = s.format
	}

	return out
}

func TestParse_Deleted(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []blockKind{blockDeleted}, kinds(Parse("[deleted]")))
}

func TestParse_ParagraphSplitting(t *testing.T) {
	t.Parallel()

	t.Run("p tags separate paragraphs", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("first<p>second")
		require.Equal(t, []blockKind{blockParagraph, blockParagraph}, kinds(blocks))
		assert.Equal(t, []string{"first"}, spanTexts(blocks[0]))
		assert.Equal(t, []string{"second"}, spanTexts(blocks[1]))
	})

	t.Run("leading p is a no-op", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("<p>first<p>second")
		assert.Equal(t, []blockKind{blockParagraph, blockParagraph}, kinds(blocks))
	})

	t.Run("empty paragraphs vanish", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("first<p><p>second")
		require.Equal(t, []blockKind{blockParagraph, blockParagraph}, kinds(blocks))
	})

	t.Run("entities decode", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("a &amp; b&#x27;s")
		assert.Equal(t, []string{"a & b's"}, spanTexts(blocks[0]))
	})
}

func TestParse_QuoteClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		html string
		want string
	}{
		{"plain marker", "&gt; quoted here", "quoted here"},
		{"leading space", " &gt; quoted here", "quoted here"},
		{"double marker", "&gt;&gt; quoted here", "quoted here"},
		{"italic wrapped", "<i>&gt; quoted here</i>", "quoted here"},
		{"italic with space", "<i> &gt; quoted here</i>", "quoted here"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			blocks := Parse(tt.html)
			require.Equal(t, []blockKind{blockQuote}, kinds(blocks))
			assert.Equal(t, []string{tt.want}, spanTexts(blocks[0]))
		})
	}

	t.Run("mid-text marker is no quote", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []blockKind{blockParagraph}, kinds(Parse("this &gt; is not a quote")))
	})
}

func TestParse_CodeBlocks(t *testing.T) {
	t.Parallel()

	t.Run("code with following text splits", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("intro:<p><pre><code>  x := 1\n</code></pre>\nafter text")
		require.Equal(t, []blockKind{blockParagraph, blockCode, blockParagraph}, kinds(blocks))
		assert.Equal(t, "  x := 1\n", blocks[1].text)
		assert.Equal(t, []string{"after text"}, spanTexts(blocks[2]))
	})

	t.Run("p directly after code adds no empty block", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("intro:<p><pre><code>  x\n</code></pre><p>closing")
		assert.Equal(t, []blockKind{blockParagraph, blockCode, blockParagraph}, kinds(blocks))
	})

	t.Run("input ending at code has no trailing block", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("intro:<p><pre><code>  x\n</code></pre>")
		assert.Equal(t, []blockKind{blockParagraph, blockCode}, kinds(blocks))
	})
}

func TestParse_Anchors(t *testing.T) {
	t.Parallel()

	t.Run("truncated display restores to href", func(t *testing.T) {
		t.Parallel()

		blocks := Parse(`see <a href="https:&#x2F;&#x2F;example.com&#x2F;full&#x2F;path">https:&#x2F;&#x2F;example.com&#x2F;fu...</a> here`)
		require.Equal(t, []spanFormat{spanPlain, spanLink, spanPlain}, spanFormats(blocks[0]))
		assert.Equal(t, "https://example.com/full/path", blocks[0].spans[1].href)
	})

	t.Run("sentence period after anchor survives", func(t *testing.T) {
		t.Parallel()

		blocks := Parse(`see <a href="https:&#x2F;&#x2F;example.com">https:&#x2F;&#x2F;example.com</a>. next`)
		require.Equal(t, []spanFormat{spanPlain, spanLink, spanPlain}, spanFormats(blocks[0]))
		assert.Equal(t, ". next", blocks[0].spans[2].text)
	})

	t.Run("href with commas stays one link", func(t *testing.T) {
		t.Parallel()

		blocks := Parse(`<a href="https:&#x2F;&#x2F;example.com&#x2F;a,b">https:&#x2F;&#x2F;example.com&#x2F;a,b</a> x`)
		require.Equal(t, []spanFormat{spanLink, spanPlain}, spanFormats(blocks[0]))
		assert.Equal(t, "https://example.com/a,b", blocks[0].spans[0].href)
	})

	t.Run("bare URL sheds its sentence period", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("see https:&#x2F;&#x2F;example.com&#x2F;x. Next sentence")
		require.Equal(t, []spanFormat{spanPlain, spanLink, spanPlain}, spanFormats(blocks[0]))
		assert.Equal(t, "https://example.com/x", blocks[0].spans[1].href)
		assert.Equal(t, ". Next sentence", blocks[0].spans[2].text)
	})
}

func TestTokenize_BacktickPairs(t *testing.T) {
	t.Parallel()

	t.Run("even count marks code", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("use `go vet` always")
		assert.Equal(t, []spanFormat{spanPlain, spanCodeInline, spanPlain}, spanFormats(blocks[0]))
		assert.Equal(t, []string{"use ", "go vet", " always"}, spanTexts(blocks[0]))
	})

	t.Run("odd count leaves text alone", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("a stray ` backtick")
		assert.Equal(t, []spanFormat{spanPlain}, spanFormats(blocks[0]))
	})
}

func TestTokenize_VariablesSuppressedByStrayBacktick(t *testing.T) {
	t.Parallel()

	t.Run("variables highlight normally", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("the $HOME variable")
		assert.Equal(t, []spanFormat{spanPlain, spanVariable, spanPlain}, spanFormats(blocks[0]))
	})

	t.Run("stray backtick suppresses variables", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("a ` stray and $HOME")
		assert.Equal(t, []spanFormat{spanPlain}, spanFormats(blocks[0]))
	})

	t.Run("consumed pairs let variables through", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("`code` and $HOME")
		assert.Equal(t,
			[]spanFormat{spanCodeInline, spanPlain, spanVariable},
			spanFormats(blocks[0]))
	})
}

func TestTokenize_MentionsAndReferences(t *testing.T) {
	t.Parallel()

	blocks := Parse("thanks @someone see [1] and [10]")

	assert.Equal(t, []spanFormat{
		spanPlain, spanMention, spanPlain, spanReference, spanPlain, spanReference,
	}, spanFormats(blocks[0]))

	// The historical regex styles the leading space with the handle.
	assert.Equal(t, " @someone", blocks[0].spans[1].text)
	assert.Equal(t, "1", blocks[0].spans[3].text)
	assert.Equal(t, "10", blocks[0].spans[5].text)
}

func TestTokenize_QuotesOnlyGetLinks(t *testing.T) {
	t.Parallel()

	blocks := Parse("&gt; quoting `code` and $VAR and https:&#x2F;&#x2F;example.com&#x2F;x here")

	require.Equal(t, []blockKind{blockQuote}, kinds(blocks))
	assert.Equal(t, []spanFormat{spanPlain, spanLink, spanPlain}, spanFormats(blocks[0]))
}

func TestTypography_Symbols(t *testing.T) {
	t.Parallel()

	blocks := Parse("wait... CO2 rose 1/2 -- twice--fast")

	assert.Equal(t, []string{"wait… CO₂ rose ½ — twice—fast"}, spanTexts(blocks[0]))
}

func TestTypography_CO2NeedsWordBoundaries(t *testing.T) {
	t.Parallel()

	t.Run("inside a word stays put", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("the ACO2 sensor, an MCO2X part, CO2e estimates")
		assert.Equal(t, []string{"the ACO2 sensor, an MCO2X part, CO2e estimates"}, spanTexts(blocks[0]))
	})

	t.Run("against punctuation converts", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("(CO2) and CO2-neutral and CO2.")
		assert.Equal(t, []string{"(CO₂) and CO₂-neutral and CO₂."}, spanTexts(blocks[0]))
	})
}

func TestTypography_SmileyNeedsWhitespace(t *testing.T) {
	t.Parallel()

	t.Run("after space converts", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("nice :)")
		assert.Equal(t, []string{"nice 😊"}, spanTexts(blocks[0]))
	})

	t.Run("glued stays put", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("nice:)")
		assert.Equal(t, []string{"nice:)"}, spanTexts(blocks[0]))
	})

	t.Run("whole comment converts", func(t *testing.T) {
		t.Parallel()

		blocks := Parse(":)")
		assert.Equal(t, []string{"😊"}, spanTexts(blocks[0]))
	})

	t.Run("glued to a word stays put", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("hi to :Dave and :Python and a :/etc path")
		assert.Equal(t, []string{"hi to :Dave and :Python and a :/etc path"}, spanTexts(blocks[0]))
	})

	t.Run("before punctuation converts", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("fun :D, right :)")
		assert.Equal(t, []string{"fun 😄, right 😊"}, spanTexts(blocks[0]))
	})

	t.Run("consecutive smileys share their boundary spaces", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("great :) :)")
		assert.Equal(t, []string{"great 😊 😊"}, spanTexts(blocks[0]))
	})
}

func TestTokenize_InsideItalics(t *testing.T) {
	t.Parallel()

	t.Run("backtick pairs in italics become code, not literal ticks", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("<i>use `go vet` here</i>")
		assert.Equal(t, []spanFormat{spanItalic, spanCodeInline, spanItalic}, spanFormats(blocks[0]))
		assert.NotContains(t, blocks[0].spans[1].text, "`")
	})

	t.Run("consumed italic backticks do not suppress variables", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("<i>`x`</i> and $foo")
		assert.Contains(t, spanFormats(blocks[0]), spanVariable)
	})

	t.Run("tokens keep their role, remainder keeps italics", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("<i>see [1] and (YC W21) here</i>")
		assert.Equal(t,
			[]spanFormat{spanItalic, spanReference, spanItalic, spanYCLabel, spanItalic},
			spanFormats(blocks[0]))
	})
}

func TestParse_OperatorLeadIsNotAQuote(t *testing.T) {
	t.Parallel()

	t.Run("leading >= stays a verbatim paragraph", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("&gt;= 3 versions are affected")
		require.Equal(t, []blockKind{blockParagraph}, kinds(blocks))
		assert.Equal(t, []string{">= 3 versions are affected"}, spanTexts(blocks[0]))
	})

	t.Run("leading >>= stays a verbatim paragraph", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("&gt;&gt;= sequences monadic actions")
		require.Equal(t, []blockKind{blockParagraph}, kinds(blocks))
		assert.Equal(t, []string{">>= sequences monadic actions"}, spanTexts(blocks[0]))
	})

	t.Run("quoted operator text keeps its characters", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("&gt; &gt;= 3 versions are affected")
		require.Equal(t, []blockKind{blockParagraph}, kinds(blocks))
		assert.Equal(t, []string{"> >= 3 versions are affected"}, spanTexts(blocks[0]))
	})

	t.Run("plain quotes still convert", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("&gt; a normal quote")
		require.Equal(t, []blockKind{blockQuote}, kinds(blocks))
		assert.Equal(t, []string{"a normal quote"}, spanTexts(blocks[0]))
	})
}

func TestParse_QuoteFoldsItalics(t *testing.T) {
	t.Parallel()

	blocks := Parse("&gt; a quote with <i>a URL https:&#x2F;&#x2F;example.org&#x2F;x inside</i> here")

	require.Equal(t, []blockKind{blockQuote}, kinds(blocks))
	assert.Equal(t, []spanFormat{spanPlain, spanLink, spanPlain}, spanFormats(blocks[0]),
		"italics dissolve in quotes, so the URL inside must still be linkified")
}

func TestParse_BlockEdgeWhitespaceTrims(t *testing.T) {
	t.Parallel()

	blocks := Parse("a<p><pre><code>  x\n</code></pre><p>\nfoo")

	require.Equal(t, []blockKind{blockParagraph, blockCode, blockParagraph}, kinds(blocks))
	assert.Equal(t, []string{"foo"}, spanTexts(blocks[2]),
		"formatting whitespace at block edges trims away")
}

func TestTypography_SmileyRunsBeforeQuoteStrip(t *testing.T) {
	t.Parallel()

	t.Run("marker glued to smiley blocks conversion", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("&gt;:)")
		assert.Equal(t, []string{":)"}, spanTexts(blocks[0]))
	})

	t.Run("leading space still converts", func(t *testing.T) {
		t.Parallel()

		blocks := Parse(" :) yes")
		assert.Equal(t, []string{"😊 yes"}, spanTexts(blocks[0]))
	})
}

func TestNormalize_NewlineJoins(t *testing.T) {
	t.Parallel()

	t.Run("soft break joins", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("line one\nline two")
		assert.Equal(t, []string{"line one line two"}, spanTexts(blocks[0]))
	})

	t.Run("break before opening bracket survives", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("line one\n(aside)")
		assert.Equal(t, []string{"line one\n(aside)"}, spanTexts(blocks[0]))
	})

	t.Run("spaces collapse", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("a  b   c")
		assert.Equal(t, []string{"a b c"}, spanTexts(blocks[0]))
	})

	t.Run("newline run becomes one hard break", func(t *testing.T) {
		t.Parallel()

		blocks := Parse("line one\n\n\nline two")
		assert.Equal(t, []string{"line one\nline two"}, spanTexts(blocks[0]))
	})
}
