package headline

import (
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/nerdfonts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		title string
		want  []span
	}{
		{
			"plain title",
			"Nothing special here",
			[]span{{text: "Nothing special here", kind: spanPlain}},
		},
		{
			"prefix at start",
			"Ask HN: How do I grow?",
			[]span{{text: "Ask HN:", kind: spanHNPrefix}, {text: " How do I grow?", kind: spanPlain}},
		},
		{
			"everything at once",
			"Show HN: Foo (YC W05) from (2019) [pdf]",
			[]span{
				{text: "Show HN:", kind: spanHNPrefix},
				{text: " Foo ", kind: spanPlain},
				{text: "YC W05", kind: spanYCLabel},
				{text: " from ", kind: spanPlain},
				{text: "2019", kind: spanYear},
				{text: " ", kind: spanPlain},
				{text: "pdf", kind: spanContentTag},
			},
		},
		{
			"uppercase tag",
			"Annual report [PDF]",
			[]span{{text: "Annual report ", kind: spanPlain}, {text: "PDF", kind: spanContentTag}},
		},
		{
			"yc label is not a year and vice versa",
			"Startup (YC X25) raised in (2024)",
			[]span{
				{text: "Startup ", kind: spanPlain},
				{text: "YC X25", kind: spanYCLabel},
				{text: " raised in ", kind: spanPlain},
				{text: "2024", kind: spanYear},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, parse(tc.title))
		})
	}
}

func TestParse_PrefixOnlyAtStart(t *testing.T) {
	t.Parallel()

	spans := parse(`Tell HN: why "Show HN:" posts fail`)

	require.Len(t, spans, 2)
	assert.Equal(t, span{text: "Tell HN:", kind: spanHNPrefix}, spans[0])
	assert.Equal(t, spanPlain, spans[1].kind, `a quoted "Show HN:" mid-title stays prose`)
}

// A title that genuinely contains a content glyph must stay plain — only the
// bracketed source form is a tag. This is the marker-glyph round-trip bug the
// span rewrite eliminates.
func TestParse_GlyphIsNotATag(t *testing.T) {
	t.Parallel()

	title := "Weird unicode " + nerdfonts.Document + " in a title"

	assert.Equal(t, []span{{text: title, kind: spanPlain}}, parse(title))
}

var allStates = []HighlightType{
	Unselected, HeadlineInCommentSection, Selected, OpenStory, MarkAsRead, AddToFavorites, RemoveFromFavorites,
}

func TestRender_DisplayText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		title string
		want  string
	}{
		{"Ask HN: What now?", "Ask HN: What now?"},
		{"Foo (YC W05) bar (2019)", "Foo YC W05 bar 2019"},
		{"Report [pdf] out", "Report pdf out"},
		{"Plain", "Plain"},
	}

	for _, tc := range cases {
		for _, state := range allStates {
			assert.Equal(t, tc.want, ansi.Strip(Render(tc.title, state, false)),
				"visible text for %q in state %d", tc.title, state)
		}
	}
}

func TestRender_UnselectedPlainTitleIsUntouched(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "just words", Render("just words", Unselected, false))
}

// The desync regression: with the old spliced-ANSI pipeline, the raw escapes
// that reopened the row style after a token had to mirror the list's styles
// by hand — a mismatch broke every cell after the token. Composition makes
// the base attributes hold across the whole title, tokens included.
func TestRender_BaseAttributesHoldAcrossTokens(t *testing.T) {
	t.Parallel()

	title := "Ask HN: Foo (YC W05) bar (2019) [pdf] tail"

	for cell, a := range cellAttrs(Render(title, Selected, false)) {
		assert.True(t, a.reverse, "Selected: cell %d must be reversed", cell)
	}

	for cell, a := range cellAttrs(Render(title, MarkAsRead, false)) {
		assert.True(t, a.faint, "MarkAsRead: cell %d must be faint", cell)
		assert.True(t, a.italic, "MarkAsRead: cell %d must be italic", cell)
	}

	for cell, a := range cellAttrs(Render(title, HeadlineInCommentSection, false)) {
		assert.True(t, a.bold, "comment-section header: cell %d must be bold", cell)
	}

	for cell, a := range cellAttrs(Render(title, AddToFavorites, false)) {
		assert.True(t, a.reverse, "favorites flash: cell %d must keep the reversed bar", cell)
	}

	for cell, a := range cellAttrs(Render(title, OpenStory, false)) {
		assert.True(t, a.brightBlackBg, "open story: cell %d must sit on the muted bar", cell)
		assert.False(t, a.reverse, "open story: cell %d must not fall back to reverse video", cell)
	}
}

func TestRender_NerdFonts(t *testing.T) {
	t.Parallel()

	tag := Render("Report [pdf]", Unselected, true)
	assert.Contains(t, tag, nerdfonts.Document, "the tag renders as its glyph")
	assert.NotContains(t, ansi.Strip(tag), "[pdf]")

	pill := Render("Foo (YC W05)", Selected, true)
	assert.Contains(t, pill, nerdfonts.LeftSeparator)
	assert.Contains(t, pill, nerdfonts.RightSeparator)
	assert.Contains(t, pill, nerdfonts.YCombinator)
	assert.Contains(t, ansi.Strip(pill), "W05")
	assert.NotContains(t, ansi.Strip(pill), "YC W05", "the pill shows the glyph plus season, not the words")
}

func TestHighlightDomain(t *testing.T) {
	t.Parallel()

	assert.Empty(t, HighlightDomain(""))
	assert.Equal(t, "(example.com)", ansi.Strip(HighlightDomain("example.com")))
	assert.Contains(t, HighlightDomain("example.com"), ansi.Faint)
}

type attrs struct {
	bold, faint, italic, reverse, brightBlackBg bool
}

var sgrSeq = regexp.MustCompile("^\x1b\\[([0-9;:]*)m")

// cellAttrs interprets s as a terminal would, mapping each printable rune's
// starting cell to the SGR attributes active when it is drawn. Color
// parameters (38/48/58 forms) are skipped so their arguments are not
// mistaken for attribute codes.
func cellAttrs(s string) map[int]attrs {
	cells := make(map[int]attrs)

	var a attrs

	cell := 0

	for len(s) > 0 {
		if m := sgrSeq.FindStringSubmatch(s); m != nil {
			applySGR(&a, m[1])
			s = s[len(m[0]):]

			continue
		}

		_, size := utf8.DecodeRuneInString(s)
		cells[cell] = a
		cell++
		s = s[size:]
	}

	return cells
}

func applySGR(a *attrs, params string) {
	parts := strings.Split(params, ";")

	for i := 0; i < len(parts); i++ {
		p := parts[i]

		if strings.HasPrefix(p, "38:") || strings.HasPrefix(p, "48:") || strings.HasPrefix(p, "58:") {
			continue
		}

		switch p {
		case "", "0":
			*a = attrs{}
		case "1":
			a.bold = true
		case "2":
			a.faint = true
		case "3":
			a.italic = true
		case "7":
			a.reverse = true
		case "22":
			a.bold, a.faint = false, false
		case "23":
			a.italic = false
		case "27":
			a.reverse = false
		case "100":
			a.brightBlackBg = true
		case "49":
			a.brightBlackBg = false
		case "38", "48", "58":
			// Semicolon color form: skip the color arguments.
			if i+1 < len(parts) && parts[i+1] == "2" {
				i += 4
			} else if i+1 < len(parts) && parts[i+1] == "5" {
				i += 2
			}
		}
	}
}
