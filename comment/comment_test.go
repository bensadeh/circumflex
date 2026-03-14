package comment

import (
	"clx/settings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestGetParagraphSeparator(t *testing.T) {
	t.Parallel()

	t.Run("last paragraph returns empty", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, getParagraphSeparator(2, 3))
	})

	t.Run("single paragraph returns empty", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, getParagraphSeparator(0, 1))
	})

	t.Run("non-last paragraph returns double newline", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "\n\n", getParagraphSeparator(0, 3))
		assert.Equal(t, "\n\n", getParagraphSeparator(1, 3))
	})
}

func defaultConfig() *settings.Config {
	return &settings.Config{
		CommentWidth:      70,
		IndentationSymbol: " \u258e",
	}
}

func TestPrintDeleted(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	result := Print("[deleted]", cfg, 70, 80)

	assert.Contains(t, result, "[deleted]")
	assert.Contains(t, result, "\033[2m", "should contain faint ANSI escape")
}

func TestPrintSimpleText(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	result := Print("Hello &amp; world", cfg, 70, 80)

	assert.Contains(t, result, "Hello & world")
	assert.NotContains(t, result, "&amp;")
}

func TestPrintCodeBlock(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	input := "<pre><code>fmt.Println(\"hello\")\n</code></pre>"
	result := Print(input, cfg, 70, 80)

	assert.Contains(t, result, dimmed, "code block should contain dimmed ANSI")
	assert.Contains(t, result, reset, "code block should contain reset ANSI")
	assert.Contains(t, result, "fmt.Println")
}

func TestPrintQuoteBlock(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()

	// HN API wraps each paragraph with <p>. The first <p> is stripped,
	// subsequent <p> tags split into separate paragraphs.
	input := "<p>intro<p>>This is quoted"
	result := Print(input, cfg, 70, 80)

	assert.Contains(t, result, italic, "quote should contain italic ANSI")
	assert.Contains(t, result, dimmed, "quote should contain dimmed ANSI")
	assert.Contains(t, result, "This is quoted")
}

func TestPrintDisableEmojis(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	cfg.DisableEmojis = true

	result := Print("hello :)", cfg, 70, 80)

	assert.Contains(t, result, ":)")
	assert.NotContains(t, result, "\U0001f60a")
}

func TestPrintEnableEmojis(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	cfg.DisableEmojis = false

	result := Print("hello :)", cfg, 70, 80)

	assert.NotContains(t, result, ":)")
}

func TestPrintDisableCommentHighlighting(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	cfg.DisableCommentHighlighting = true

	input := "check `code` here"
	result := Print(input, cfg, 70, 80)

	assert.Contains(t, result, "`code`", "backticks should remain unhighlighted")
}

func TestPrintEnableCommentHighlighting(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	cfg.DisableCommentHighlighting = false

	input := "check `code` here"
	result := Print(input, cfg, 70, 80)

	// When highlighting is enabled, backticks are replaced with ANSI styling.
	assert.NotContains(t, result, "`code`")
}

func TestPrintMultipleParagraphs(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()

	// HN API prefixes each paragraph with <p>. The first is stripped;
	// the second acts as the paragraph separator.
	input := "<p>first paragraph<p>second paragraph"
	result := Print(input, cfg, 70, 80)

	assert.Contains(t, result, "first paragraph")
	assert.Contains(t, result, "second paragraph")
	assert.Contains(t, result, "\n\n", "paragraphs should be separated by double newline")
}
