package comment

import (
	"testing"

	"github.com/bensadeh/circumflex/ansi"

	"github.com/bensadeh/circumflex/style"
)

// FuzzParse asserts the pipeline never panics and never emits a live escape,
// whatever bytes arrive. The seeds cover every construct HN emits plus the
// shapes that historically held bugs; `go test` runs the seeds,
// `go test -fuzz=FuzzParse ./comment/` explores from them.
func FuzzParse(f *testing.F) {
	seeds := []string{
		"",
		"<p>",
		"[deleted]",
		"[flagged]",
		"plain text with &amp; entities &#x27;here&#x27;",
		"&gt; a quote<p>reply<p>&gt;&gt; deeper",
		"<i>&gt; italic quote with <a href=\"https://x.com/a,b\">https://x.com/a,b</a></i>",
		"code:<p><pre><code>  x := 1\n\n  y\n</code></pre>\nafter<p>end",
		"`ticks` and $VAR and @user and [1] and (YC W21) and IANAL :) 1/2 CO2 a--b ...",
		"<i>nested <i>italics</i> with `code` inside</i> tail",
		"line\n\nbreaks\n[1] footnote\n(aside)",
		"<a href=\"\">empty</a><a>no href</a><pre></pre><p><p>",
		"https://example.com/x. Sentence https://example.com/y",
		"\x00\x01\x1b[31mraw bytes�",
		"&#27;[31mentity escape&#x9b;0;t&#7;",
		"<a href=\"https://x.com/&#27;]8;;osc\">breakout</a>",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		blocks := Parse(input)

		// Parse owns sanitation: the ingestion strip ran before entities
		// decoded, so whatever arrives — raw escape bytes or entity encodings
		// of them — no parsed text may hold a live escape.
		for _, b := range blocks {
			assertSanitized(t, b.text)

			for _, s := range b.spans {
				assertSanitized(t, s.text)
				assertSanitized(t, s.href)
			}
		}

		for _, opts := range []RenderOptions{
			{CommentWidth: 72, ScreenWidth: 80},
			{CommentWidth: 40, ScreenWidth: 44, NerdFonts: true, Fg: style.CommentModFg()},
			{CommentWidth: 10, ScreenWidth: 12},
		} {
			_ = RenderBlocks(blocks, opts)
			_ = RenderContent(blocks, 3, opts)
		}
	})
}

func assertSanitized(t *testing.T, s string) {
	t.Helper()

	if stripped := ansi.Strip(s); stripped != s {
		t.Fatalf("unsanitized parse output %q", s)
	}
}
