package comment

import (
	"image/color"
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/style"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintDeleted(t *testing.T) {
	t.Parallel()

	result := renderBody("[deleted]", 70, 80, false, nil)

	assert.Contains(t, result, "[deleted]")
	assert.Contains(t, result, "\033[2m", "should contain faint ANSI escape")
}

func TestPrintSimpleText(t *testing.T) {
	t.Parallel()

	result := renderBody("Hello &amp; world", 70, 80, false, nil)

	assert.Contains(t, result, "Hello & world")
	assert.NotContains(t, result, "&amp;")
}

func TestPrintCodeBlock(t *testing.T) {
	t.Parallel()

	input := "<pre><code>fmt.Println(\"hello\")\n</code></pre>"
	result := renderBody(input, 70, 80, false, nil)

	assert.Contains(t, result, ansi.Faint, "code block should contain dimmed ANSI")
	assert.Contains(t, result, ansi.Reset, "code block should contain reset ANSI")
	assert.Contains(t, result, "fmt.Println")
}

func TestPrintQuoteBlock(t *testing.T) {
	t.Parallel()

	// HN API wraps each paragraph with <p>. The first <p> is stripped,
	// subsequent <p> tags split into separate paragraphs.
	input := "<p>intro<p>>This is quoted"
	result := renderBody(input, 70, 80, false, nil)

	assert.Contains(t, result, ansi.Italic, "quote should contain italic ANSI")
	assert.Contains(t, result, ansi.Faint, "quote should contain dimmed ANSI")
	assert.Contains(t, result, "This is quoted")
}

func TestPrintConvertsSmileys(t *testing.T) {
	t.Parallel()

	result := renderBody("hello :)", 70, 80, false, nil)

	assert.NotContains(t, result, ":)")
}

func TestPrintCommentHighlighting(t *testing.T) {
	t.Parallel()

	input := "check `code` here"
	result := renderBody(input, 70, 80, false, nil)

	assert.NotContains(t, result, "`code`")
}

func TestPrintMultipleParagraphs(t *testing.T) {
	t.Parallel()

	// HN API prefixes each paragraph with <p>. The first is stripped;
	// the second acts as the paragraph separator.
	input := "<p>first paragraph<p>second paragraph"
	result := renderBody(input, 70, 80, false, nil)

	assert.Contains(t, result, "first paragraph")
	assert.Contains(t, result, "second paragraph")
	assert.Contains(t, result, "\n\n", "paragraphs should be separated by double newline")
}

// tintParams returns the SGR parameter string of a color's foreground
// escape, e.g. "38;5;131".
func tintParams(t *testing.T, c color.Color) string {
	t.Helper()

	code := style.ForegroundCode(c)
	require.True(t, strings.HasPrefix(code, "\x1b[") && strings.HasSuffix(code, "m"),
		"expected mod color to render a foreground escape, got %q", code)

	return strings.TrimSuffix(strings.TrimPrefix(code, "\x1b["), "m")
}

// assertTintedEverywhere walks the rendered bytes tracking SGR state and
// fails if any visible character is reached while the tint foreground is
// inactive. This is the semantic contract of the mod tint: no matter how
// spans reset styling and wrapping re-opens it, the color holds wherever
// text shows.
func assertTintedEverywhere(t *testing.T, rendered, fgParams string) {
	t.Helper()

	active := false

	for i := 0; i < len(rendered); {
		if rendered[i] == '\n' {
			i++

			continue
		}

		if rendered[i] != '\x1b' {
			assert.Truef(t, active, "visible byte %q at offset %d rendered without the tint", rendered[i], i)
			i++

			continue
		}

		rest := rendered[i:]

		switch {
		case strings.HasPrefix(rest, "\x1b["): // SGR
			end := strings.IndexByte(rest, 'm')
			require.GreaterOrEqual(t, end, 0, "unterminated SGR sequence at offset %d", i)

			params := rest[2:end]
			if params == "" || params == "0" {
				active = false
			} else if strings.Contains(params, fgParams) {
				active = true
			}

			i += end + 1

		case strings.HasPrefix(rest, "\x1b]"): // OSC (hyperlinks), ends with BEL
			end := strings.IndexByte(rest, '\a')
			require.GreaterOrEqual(t, end, 0, "unterminated OSC sequence at offset %d", i)

			i += end + 1

		default:
			i++
		}
	}
}

func TestRender_ModParagraphTintSurvivesStyledSpans(t *testing.T) {
	t.Parallel()

	// Four tint-killing span kinds in one paragraph: inline code, mention,
	// hyperlink, italics. Every visible character must still carry the tint.
	input := "see `code`, @user, <a href=\"https://example.com\">https://example.com</a>, and <i>italic</i> here"

	result := renderBody(input, 80, 80, false, style.CommentModFg())

	assertTintedEverywhere(t, result, tintParams(t, style.CommentModFg()))
}

func TestRender_ModParagraphTintSurvivesWrapping(t *testing.T) {
	t.Parallel()

	// Long enough to wrap onto multiple lines at width 40; the tint must be
	// re-established on every wrapped line.
	input := strings.Repeat("word ", 20)

	result := renderBody(input, 40, 40, false, style.CommentModFg())

	lines := strings.Split(result, "\n")
	require.Greater(t, len(lines), 1, "test input should wrap to multiple lines")

	assertTintedEverywhere(t, result, tintParams(t, style.CommentModFg()))
}

func TestRender_QuoteStyleSurvivesEmbeddedLink(t *testing.T) {
	t.Parallel()

	// Text after a link inside a quote must return to the quote's faint
	// italic instead of rendering plain.
	input := "&gt; before <a href=\"https://example.com/x\">https://example.com/x</a> after"

	result := renderBody(input, 80, 80, false, nil)

	idx := strings.LastIndex(result, "after")
	require.Positive(t, idx, "quote text should contain the trailing words")

	head := result[:idx]
	assert.Greater(t, strings.LastIndex(head, ansi.Italic+ansi.Faint), strings.LastIndex(head, "\x1b[m"),
		"the faint italic base must be re-opened after the link's reset")
}
