package article

import (
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/ansi"

	"golang.org/x/net/html"
)

// FuzzParseBlocks pins the reader's half of the sanitation contract: a page
// is wholly attacker-controlled, so no text the parser hands the renderer may
// carry a live escape, whether it arrived as a text node, an attribute, a
// table cell or an entity that only decodes into one late. The pipeline runs
// the passes extractReadable runs, minus readability itself — readability
// only deletes nodes, so parsing without it sees strictly more of them.
func FuzzParseBlocks(f *testing.F) {
	seeds := []string{
		"",
		"<p>plain paragraph</p>",
		"<p>\x1b[31mraw escape\x1b[0m</p>",
		"<p>&#27;[31mentity escape&#x9b;0;t&#7;</p>",
		"<img src=\"https://x.com/a.png\" alt=\"\x1b]0;pwn\x07caption\">",
		"<div role=\"img\" aria-label=\"chart \x1b[31m\"><svg><circle></circle></svg></div>",
		"<svg role=\"img\" aria-label=\"\x1b]8;;https://evil.com\x07plot\"></svg>",
		"<table><tr><th>a\x1b[7m</th><td>b</td></tr></table>",
		"<pre><code class=\"language-go\x1b[31m\">x := 1\n\ty\n</code></pre>",
		"<pre><span class=\"comment\" id=\"x.Related\">// c</span><a href=\"https://x.com/\x1b]8;;\">t</a></pre>",
		"<a href=\"https://x.com/\x07evil\">link</a><a href=\"javascript:x\">no</a>",
		"<h1>heading\rrewrite</h1><blockquote>quoted</blockquote>",
		"<ul><li>one \x1b(0 two</li><li>three</li></ul>",
		"<math alttext=\"x^2 \x1b[0m\"><mi>x</mi></math>",
		"<img src=\"https://latex.codecogs.com/png?\x1b[31m\" alt=\"\">",
		"text with \xc2\x11\x9b invalid utf-8",
		"<p>line\r\nbreaks\rand a lone carriage return</p>",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		doc, err := html.Parse(strings.NewReader(input))
		if err != nil {
			return
		}

		normalizeMediaWiki(doc)
		normalizeRoleImages(doc)
		preserveCodeLang(doc)
		preservePreContent(doc)
		normalizeLatexmlTables(doc)
		dropLatexmlRawPictures(doc)

		blocks := parseBlocks(doc)
		convertMath(blocks)
		guessCodeLangs(blocks)

		assertBlocksSanitized(t, applySiteRules(blocks, "example.com"))
		assertBlocksSanitized(t, parseTextBlocks(input))
	})
}

func assertBlocksSanitized(t *testing.T, blocks []block) {
	t.Helper()

	for _, b := range blocks {
		assertSanitized(t, b.text)
		assertSanitized(t, b.lang)
		assertSpansSanitized(t, b.spans)

		for _, item := range b.items {
			assertSpansSanitized(t, item.spans)
		}

		for _, row := range b.rows {
			for _, cell := range row {
				assertSanitized(t, cell)
			}
		}
	}
}

func assertSpansSanitized(t *testing.T, spans []span) {
	t.Helper()

	for _, s := range spans {
		assertSanitized(t, s.text)
		assertSanitized(t, s.href)
	}
}

func assertSanitized(t *testing.T, s string) {
	t.Helper()

	if stripped := ansi.Strip(s); stripped != s {
		t.Fatalf("unsanitized parse output %q", s)
	}
}
