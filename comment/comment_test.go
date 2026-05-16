package comment

import (
	"strings"
	"testing"

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

func TestRender_ModParagraphTintReappliedAfterEveryReset(t *testing.T) {
	t.Parallel()

	// Each ansi.Reset emitted by a styled span must be followed by the mod-tint
	// foreground escape — otherwise plaintext after the span silently loses tint.
	// Exercises four reset sources: inline code, mention, URL hyperlink, <i> tag.
	input := "see `code`, @user, <a href=\"https://example.com\">link</a>, and <i>italic</i> here"

	result := Render(input, 80, 80, false, style.CommentModFg())

	require.Contains(t, result, ansi.Reset, "test input should produce styled spans with resets")

	for i := 0; i < len(result); {
		rel := strings.Index(result[i:], ansi.Reset)
		if rel == -1 {
			break
		}

		pos := i + rel
		after := pos + len(ansi.Reset)
		i = after

		if after >= len(result) {
			continue
		}

		assert.Equalf(t, byte(0x1B), result[after],
			"reset at offset %d must be followed by ESC (color re-apply), got %q",
			pos, string(result[after]))
	}
}
