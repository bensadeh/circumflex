package comment

import (
	"testing"

	"github.com/bensadeh/circumflex/style"
)

// FuzzParse asserts the pipeline never panics, whatever bytes arrive. The
// seeds cover every construct HN emits plus the shapes that historically
// held bugs; `go test` runs the seeds, `go test -fuzz=FuzzParse ./comment/`
// explores from them.
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
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		blocks := Parse(input)

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
