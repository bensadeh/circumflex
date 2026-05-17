package comment

import (
	"image/color"
	"regexp"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/style"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsQuote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want bool
	}{
		{"starts with >", ">quoted text", true},
		{"starts with space >", " >quoted text", true},
		{"starts with <i>>", "<i>>quoted text", true},
		{"starts with <i> >", "<i> >quoted text", true},
		{"plain text", "just regular text", false},
		{"empty string", "", false},
		{"contains > but not prefix", "this > is not a quote", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, isQuote(tt.text))
		})
	}
}

func TestPrintDeleted(t *testing.T) {
	t.Parallel()

	result := Render("[deleted]", 70, 80, false, nil)

	assert.Contains(t, result, "[deleted]")
	assert.Contains(t, result, "\033[2m", "should contain faint ANSI escape")
}

func TestPrintSimpleText(t *testing.T) {
	t.Parallel()

	result := Render("Hello &amp; world", 70, 80, false, nil)

	assert.Contains(t, result, "Hello & world")
	assert.NotContains(t, result, "&amp;")
}

func TestPrintCodeBlock(t *testing.T) {
	t.Parallel()

	input := "<pre><code>fmt.Println(\"hello\")\n</code></pre>"
	result := Render(input, 70, 80, false, nil)

	assert.Contains(t, result, ansi.Faint, "code block should contain dimmed ANSI")
	assert.Contains(t, result, ansi.Reset, "code block should contain reset ANSI")
	assert.Contains(t, result, "fmt.Println")
}

func TestPrintQuoteBlock(t *testing.T) {
	t.Parallel()

	// HN API wraps each paragraph with <p>. The first <p> is stripped,
	// subsequent <p> tags split into separate paragraphs.
	input := "<p>intro<p>>This is quoted"
	result := Render(input, 70, 80, false, nil)

	assert.Contains(t, result, ansi.Italic, "quote should contain italic ANSI")
	assert.Contains(t, result, ansi.Faint, "quote should contain dimmed ANSI")
	assert.Contains(t, result, "This is quoted")
}

func TestPrintConvertsSmileys(t *testing.T) {
	t.Parallel()

	result := Render("hello :)", 70, 80, false, nil)

	assert.NotContains(t, result, ":)")
}

func TestPrintCommentHighlighting(t *testing.T) {
	t.Parallel()

	input := "check `code` here"
	result := Render(input, 70, 80, false, nil)

	// Backticks are replaced with ANSI styling.
	assert.NotContains(t, result, "`code`")
}

func TestPrintMultipleParagraphs(t *testing.T) {
	t.Parallel()

	// HN API prefixes each paragraph with <p>. The first is stripped;
	// the second acts as the paragraph separator.
	input := "<p>first paragraph<p>second paragraph"
	result := Render(input, 70, 80, false, nil)

	assert.Contains(t, result, "first paragraph")
	assert.Contains(t, result, "second paragraph")
	assert.Contains(t, result, "\n\n", "paragraphs should be separated by double newline")
}

// sgrResetForTest matches both reset forms that appear in our rendered
// output: lipgloss's short form and our own long-form ansi.Reset constant.
var sgrResetForTest = regexp.MustCompile(`\x1b\[0?m`)

// expectedTintPrefix returns the raw ANSI foreground escape that
// PaintForeground reapplies after each reset.
func expectedTintPrefix(t *testing.T, c color.Color) string {
	t.Helper()

	const marker = "\xff"

	rendered := lipgloss.NewStyle().Foreground(c).Render(marker)
	idx := strings.Index(rendered, marker)
	require.Positive(t, idx, "expected mod color to render with a non-empty foreground prefix")

	return rendered[:idx]
}

func TestRender_ModParagraphTintReappliedAfterEveryReset(t *testing.T) {
	t.Parallel()

	// Every SGR reset emitted inside a mod paragraph must be immediately
	// followed by the mod-tint foreground escape — otherwise plaintext after
	// the span silently loses tint. Exercises four reset sources: inline
	// code, mention, URL hyperlink, <i> tag.
	//
	// The anchor's display text is the URL itself so that TrimURLs takes the
	// CommentURL highlighting path (it requires display text to start with
	// http(s)://); otherwise the anchor is stripped to plain text with no
	// styling and the test would lose URL-hyperlink coverage.
	//
	// Important: lipgloss emits the short form \x1b[m after styled spans;
	// our own ansi.Reset (used by syntax.ReplaceHTML) is the long form
	// \x1b[0m. Both forms must get the prefix reapplied.
	input := "see `code`, @user, <a href=\"https://example.com\">https://example.com</a>, and <i>italic</i> here"

	result := Render(input, 80, 80, false, style.CommentModFg())

	prefix := expectedTintPrefix(t, style.CommentModFg())

	resets := sgrResetForTest.FindAllStringIndex(result, -1)
	require.NotEmpty(t, resets, "test input should produce styled spans with resets")

	var shortResets, longResets int

	for _, r := range resets {
		switch r[1] - r[0] {
		case len("\x1b[m"):
			shortResets++
		case len("\x1b[0m"):
			longResets++
		}

		after := r[1]
		if after == len(result) {
			// Trailing reset (PaintForeground appends one at the end) is fine.
			continue
		}

		assert.Truef(t, strings.HasPrefix(result[after:], prefix),
			"reset at offset %d (form %q) must be followed by mod-tint prefix %q, got %q",
			r[0], result[r[0]:r[1]], prefix, snippet(result, after, len(prefix)+4))
	}

	// Guard against the test losing coverage of either reset form if upstream
	// emitters change. Both forms occur in the current rendered output.
	assert.Positive(t, shortResets, "expected at least one short-form reset \\x1b[m from lipgloss")
	assert.Positive(t, longResets, "expected at least one long-form reset \\x1b[0m from ansi.Reset")
}

// TestRender_ModParagraphLinesStartWithTint verifies the per-line prepend
// that lets the tint survive external indent-symbol prefixing (which emits
// its own reset between the indent and the comment text).
func TestRender_ModParagraphLinesStartWithTint(t *testing.T) {
	t.Parallel()

	// Long enough to wrap onto multiple lines at width 40.
	input := strings.Repeat("word ", 20)

	result := Render(input, 40, 40, false, style.CommentModFg())
	prefix := expectedTintPrefix(t, style.CommentModFg())

	lines := strings.Split(result, "\n")
	require.Greater(t, len(lines), 1, "test input should wrap to multiple lines")

	for i, line := range lines {
		if line == "" {
			continue
		}

		assert.Truef(t, strings.HasPrefix(line, prefix),
			"line %d must start with mod-tint prefix to survive external indent prefixing, got %q",
			i, snippet(line, 0, len(prefix)+4))
	}
}

// snippet returns up to n bytes of s starting at start, for test failure
// messages. Keeps assertion output short and readable.
func snippet(s string, start, n int) string {
	end := min(start+n, len(s))

	return s[start:end]
}
