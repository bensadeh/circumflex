package syntax

import (
	"strings"
	"testing"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/stretchr/testify/assert"
)

const helloWorld = "hello world"

func TestReplaceCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"apostrophe", "don&#x27;t", "don't"},
		{"greater than", "a &gt; b", "a > b"},
		{"less than", "a &lt; b", "a < b"},
		{"slash", "path&#x2F;to", "path/to"},
		{"quot entity", "&quot;hello&quot;", `"hello"`},
		{"numeric quot", "&#34;hello&#34;", `"hello"`},
		{"ampersand last", "A &amp; B", "A & B"},
		{"multiple entities", "&lt;div&gt; &amp; &quot;x&quot;", `<div> & "x"`},
		{"no entities", "plain text", "plain text"},
		{"empty string", "", ""},
		{"ampersand ordering", "&amp;gt;", "&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ReplaceCharacters(tt.input))
		})
	}
}

func TestReplaceSymbols(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ellipsis", "wait...", "wait\u2026"},
		{"CO2 subscript", "reduce CO2 now", "reduce CO\u2082 now"},
		{"double dash spaced", "hello -- world", "hello \u2014 world"},
		{"double dash inline", "hello--world", "hello\u2014world"},
		{"fraction 1/2 after space", "about 1/2 done", "about \u00bd done"},
		{"fraction 1/2 before space", "1/2 done", "\u00bd done"},
		{"fraction 1/3 after space", " 1/3", " \u2153"},
		{"fraction 2/3 after space", " 2/3", " \u2154"},
		{"fraction 1/4 after space", " 1/4", " \u00bc"},
		{"fraction 3/4 after space", " 3/4", " \u00be"},
		{"fraction 1/5 after space", " 1/5", " \u2155"},
		{"fraction 1/6 after space", " 1/6", " \u2159"},
		{"fraction 1/10 after space", " 1/10", " \u2152 "},
		{"fraction 1/5th", "1/5th", "\u2155th"},
		{"fraction 1/6th", "1/6th", "\u2159th"},
		{"fraction 1/10th", "1/10th", "\u2152 th"},
		{"no symbols", helloWorld, helloWorld},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ReplaceSymbols(tt.input))
		})
	}
}

func TestReplaceHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"strips first p tag", "<p>hello", "hello"},
		{"second p becomes newlines", "<p>first<p>second", "first\n\nsecond"},
		{"italic tags", "<i>emphasis</i>", "\033[3memphasis\033[0m"},
		{"strips closing anchor", "click</a> here", "click here"},
		{"strips pre code", "<pre><code>x</code></pre>", "x"},
		{"combined", "<p>hello<p><i>world</i></a>", "hello\n\n\033[3mworld\033[0m"},
		{"no html", "plain text", "plain text"},
		{"empty string", "", ""},
		{"multiple paragraphs", "<p>a<p>b<p>c", "a\n\nb\n\nc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ReplaceHTML(tt.input))
		})
	}
}

func TestRemoveUnwantedNewLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"joins mid-word break", "hello\nworld", helloWorld},
		{"double newline collapses first", "hello\n\nworld", "hello \nworld"},
		{"no newlines", helloWorld, helloWorld},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, RemoveUnwantedNewLines(tt.input))
		})
	}
}

func TestRemoveUnwantedWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"double space", "hello  world", helloWorld},
		{"triple space", "hello   world", helloWorld},
		{"single space unchanged", helloWorld, helloWorld},
		{"no spaces", "helloworld", "helloworld"},
		{"empty", "", ""},
		{"leading double space", "  hello", " hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, RemoveUnwantedWhitespace(tt.input))
		})
	}
}

func TestConvertSmileys(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"happy smiley", "hello :)", "hello \U0001f60a"},
		{"reverse happy", "hello (:", "hello \U0001f60a"},
		{"grin", "hello :D", "hello \U0001f604"},
		{"wink", "hello ;)", "hello \U0001f609"},
		{"tongue", "hello :P", "hello \U0001f61c"},
		{"surprised", "hello :o", "hello \U0001f62e"},
		{"sad", "hello :(", "hello \U0001f614"},
		{"unsure", "hello :/", "hello \U0001f615"},
		{"expressionless", "hello -_-", "hello \U0001f611"},
		{"neutral", "hello :|", "hello \U0001f610"},
		{"exact match", ":)", "\U0001f60a"},
		{"no whitespace before", "word:) test", "word:) test"},
		{"empty string", "", ""},
		{"no smileys", helloWorld, helloWorld},
		{"smiley with dash", "hello :-)", "hello \U0001f60a"},
		{"wink with dash", "hello ;-)", "hello \U0001f609"},
		{"unsure with dash", "hello :-/", "hello \U0001f615"},
		{"sad with dash", "hello :-(", "hello \U0001f614"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ConvertSmileys(tt.input))
		})
	}
}

func TestHighlightDomain(t *testing.T) {
	t.Run("empty domain returns reset only", func(t *testing.T) {
		result := HighlightDomain("")
		assert.Equal(t, "\033[0m", result)
	})

	t.Run("non-empty domain contains domain text", func(t *testing.T) {
		result := HighlightDomain("example.com")
		assert.Contains(t, result, "example.com")
		assert.True(t, strings.HasPrefix(result, "\033[0m"))
	})
}

func TestHighlightBackticks(t *testing.T) {
	t.Run("no backticks returns unchanged", func(t *testing.T) {
		assert.Equal(t, helloWorld, HighlightBackticks(helloWorld))
	})

	t.Run("odd backticks returns unchanged", func(t *testing.T) {
		input := "hello `world"
		assert.Equal(t, input, HighlightBackticks(input))
	})

	t.Run("single pair wraps in styled output", func(t *testing.T) {
		result := HighlightBackticks("use `code` here")
		assert.Contains(t, result, "code")
		assert.Contains(t, result, "\033[") // contains ANSI styling
		assert.NotContains(t, result, "`")
	})

	t.Run("two pairs both highlighted", func(t *testing.T) {
		result := HighlightBackticks("`a` and `b`")
		assert.Contains(t, result, "a")
		assert.Contains(t, result, "b")
		assert.NotContains(t, result, "`")
	})

	t.Run("three backticks (odd) unchanged", func(t *testing.T) {
		input := "`a` and `"
		assert.Equal(t, input, HighlightBackticks(input))
	})

	t.Run("empty string", func(t *testing.T) {
		assert.Empty(t, HighlightBackticks(""))
	})
}

func TestHighlightMentions(t *testing.T) {
	t.Run("highlights mention", func(t *testing.T) {
		result := HighlightMentions("hello @user")
		assert.Contains(t, result, "@user")
		assert.NotEqual(t, "hello @user", result)
	})

	t.Run("dang highlighted differently", func(t *testing.T) {
		result := HighlightMentions("hello @dang")
		assert.Contains(t, result, "@dang")
		// @dang should be green, not yellow
		resultOther := HighlightMentions("hello @someone")
		assert.NotEqual(t, result, strings.ReplaceAll(resultOther, "@someone", "@dang"))
	})

	t.Run("no mentions unchanged", func(t *testing.T) {
		input := helloWorld
		assert.Equal(t, input, HighlightMentions(input))
	})

	t.Run("mention at start of text", func(t *testing.T) {
		result := HighlightMentions("@user hello")
		assert.Contains(t, result, "@user")
	})

	t.Run("empty string", func(t *testing.T) {
		assert.Empty(t, HighlightMentions(""))
	})
}

func TestHighlightVariables(t *testing.T) {
	t.Run("highlights variable", func(t *testing.T) {
		result := HighlightVariables("use $PATH here")
		assert.Contains(t, result, "$PATH")
		assert.NotEqual(t, "use $PATH here", result)
	})

	t.Run("skips when backticks present", func(t *testing.T) {
		input := "use `$PATH` here"
		assert.Equal(t, input, HighlightVariables(input))
	})

	t.Run("no variables unchanged", func(t *testing.T) {
		input := helloWorld
		assert.Equal(t, input, HighlightVariables(input))
	})

	t.Run("multiple variables", func(t *testing.T) {
		result := HighlightVariables("$HOME and $USER")
		assert.Contains(t, result, "$HOME")
		assert.Contains(t, result, "$USER")
	})

	t.Run("empty string", func(t *testing.T) {
		assert.Empty(t, HighlightVariables(""))
	})
}

func TestHighlightAbbreviations(t *testing.T) {
	t.Run("IANAL highlighted in red", func(t *testing.T) {
		result := HighlightAbbreviations("but IANAL")
		assert.Contains(t, result, "IANAL")
		assert.NotEqual(t, "but IANAL", result)
	})

	t.Run("IAAL highlighted in green", func(t *testing.T) {
		result := HighlightAbbreviations("and IAAL")
		assert.Contains(t, result, "IAAL")
		assert.NotEqual(t, "and IAAL", result)
	})

	t.Run("IANAL and IAAL use different colors", func(t *testing.T) {
		rIANAL := HighlightAbbreviations("IANAL")
		rIAAL := HighlightAbbreviations("IAAL")
		// Both produce ANSI output but with different color codes
		assert.NotEqual(t, rIANAL, rIAAL)
	})

	t.Run("no abbreviations unchanged", func(t *testing.T) {
		input := helloWorld
		assert.Equal(t, input, HighlightAbbreviations(input))
	})

	t.Run("empty string", func(t *testing.T) {
		assert.Empty(t, HighlightAbbreviations(""))
	})
}

func TestHighlightReferences(t *testing.T) {
	t.Run("highlights numbered references", func(t *testing.T) {
		for i := range 11 {
			ref := "[" + strings.Repeat("", 0) + string(rune('0'+i)) + "]"
			if i == 10 {
				ref = "[10]"
			}

			result := HighlightReferences("see " + ref)
			assert.Contains(t, result, "[")
			assert.Contains(t, result, "]")
			assert.NotEqual(t, "see "+ref, result, "reference %s should be highlighted", ref)
		}
	})

	t.Run("no references unchanged", func(t *testing.T) {
		input := helloWorld
		assert.Equal(t, input, HighlightReferences(input))
	})

	t.Run("reference beyond 10 unchanged", func(t *testing.T) {
		input := "see [11]"
		assert.Equal(t, input, HighlightReferences(input))
	})

	t.Run("empty string", func(t *testing.T) {
		assert.Empty(t, HighlightReferences(""))
	})
}

func TestColorizeIndentSymbol(t *testing.T) {
	t.Run("level 0 returns reset with empty symbol", func(t *testing.T) {
		result := ColorizeIndentSymbol("|", 0)
		assert.Equal(t, "\033[0m", result)
	})

	t.Run("level 1 through 18 produce colored output", func(t *testing.T) {
		for level := 1; level <= 18; level++ {
			result := ColorizeIndentSymbol("|", level)
			assert.Contains(t, result, "|", "level %d should contain the symbol", level)
			assert.True(t, strings.HasPrefix(result, "\033[0m"), "level %d should start with reset", level)
			assert.Greater(t, len(result), len("\033[0m|"), "level %d should have ANSI codes", level)
		}
	})

	t.Run("levels 1 and 13 produce same color (cycle)", func(t *testing.T) {
		r1 := ColorizeIndentSymbol("|", 1)
		r13 := ColorizeIndentSymbol("|", 13)
		assert.Equal(t, r1, r13)
	})

	t.Run("level beyond 12 continues cycling", func(t *testing.T) {
		r7 := ColorizeIndentSymbol("|", 7)
		r19 := ColorizeIndentSymbol("|", 19)
		assert.Equal(t, r7, r19, "level 19 should wrap to same color as level 7")
	})
}

func TestTrimURLs(t *testing.T) {
	t.Run("strips HTML anchor tag", func(t *testing.T) {
		input := `<a href="https://example.com">https://example.com</a>`
		result := TrimURLs(input, true)
		assert.NotContains(t, result, "<a href")
	})

	t.Run("highlights URLs when enabled", func(t *testing.T) {
		input := "visit https://example.com/page today"
		result := TrimURLs(input, true)
		assert.Contains(t, result, "example.com/page")
		assert.NotEqual(t, input, result)
	})

	t.Run("no highlighting when disabled", func(t *testing.T) {
		input := "visit https://example.com/page today"
		result := TrimURLs(input, false)
		// With highlighting disabled, the URL text is still present
		assert.Contains(t, result, "example.com/page")
	})

	t.Run("anchor with rel attr stripped", func(t *testing.T) {
		input := `<a href="https://example.com" rel="nofollow">https://example.com`
		result := TrimURLs(input, true)
		assert.NotContains(t, result, "<a href")
		assert.Contains(t, result, "example.com")
	})

	t.Run("truncated anchor uses full href for hyperlink", func(t *testing.T) {
		input := `<a href="https://www.example.com/very/long/path/to/resource.jpg" rel="nofollow">https://www.example.com/very/long/path/to/resour…`
		result := TrimURLs(input, true)
		// Display should use the full href (scheme-stripped), not the HN-truncated text
		assert.NotContains(t, result, "resour…")
		assert.Contains(t, result, "resource.jpg")
	})

	t.Run("truncated anchor without highlight restores full URL", func(t *testing.T) {
		input := `<a href="https://www.example.com/very/long/path/to/resource.jpg" rel="nofollow">https://www.example.com/very/long/path/to/resour…`
		result := TrimURLs(input, false)
		assert.Contains(t, result, "resource.jpg")
		assert.NotContains(t, result, "resour…")
	})

	t.Run("long URL display is truncated", func(t *testing.T) {
		longPath := strings.Repeat("a", 80)
		input := "visit https://example.com/" + longPath + " today"
		result := TrimURLs(input, true)
		visible := ansi.Strip(result)
		assert.Contains(t, visible, "…")
		assert.NotContains(t, visible, longPath)
	})

	t.Run("short URL display is not truncated", func(t *testing.T) {
		input := "visit https://example.com/short today"
		result := TrimURLs(input, true)
		assert.Contains(t, result, "example.com/short")
		assert.NotContains(t, result, "…")
	})

	t.Run("non-truncated anchor preserves full URL", func(t *testing.T) {
		input := `<a href="https://example.com/page" rel="nofollow">https://example.com/page`
		result := TrimURLs(input, true)
		assert.Contains(t, result, "example.com/page")
	})

	t.Run("empty string", func(t *testing.T) {
		assert.Empty(t, TrimURLs("", false))
	})
}
