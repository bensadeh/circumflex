// Package headline styles story titles for the list rows and pane headers.
// The plain title is parsed once into typed spans — YC-batch labels, years,
// Ask/Show/Tell prefixes, content tags — and rendering composes each token's
// color with the row state's base style, so tokens inherit selection,
// read-dimming, and the rest without splicing raw escapes into styled text.
package headline

import (
	"image/color"
	"regexp"
	"strings"

	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"

	"charm.land/lipgloss/v2"
)

type HighlightType int

const (
	Unselected HighlightType = iota
	HeadlineInCommentSection
	Selected
	// OpenStory is the muted reading marker: bright-black background instead
	// of Selected's reverse video.
	OpenStory
	MarkAsRead
	AddToFavorites
	RemoveFromFavorites
)

const (
	noBreakSpace = "\u00a0"
	ansiBlack    = 16 // ANSI 256-color black
)

type spanKind int

const (
	spanPlain spanKind = iota
	spanHNPrefix
	spanYCLabel
	spanYear
	spanContentTag
)

// span is one run of the title: either plain text or a token whose text is
// already in display form (parens and brackets dropped).
type span struct {
	text string
	kind spanKind
}

var (
	reYCLabel = regexp.MustCompile(`\((YC [SWFXP]\d{2})\)`)
	reYear    = regexp.MustCompile(`\((\d{4})\)`)
)

var hnPrefixes = []struct {
	literal string
	color   func() color.Color
}{
	{"Ask HN:", style.HeadlineAskHNColor},
	{"Show HN:", style.HeadlineShowHNColor},
	{"Tell HN:", style.HeadlineTellHNColor},
	{"Thank HN:", style.HeadlineThankHNColor},
	{"Launch HN:", style.HeadlineLaunchHNColor},
}

var contentTags = []struct {
	literal string
	word    string
	glyph   string
	color   func() color.Color
}{
	{"[audio]", "audio", nerdfonts.Audio, style.HeadlineAudioColor},
	{"[video]", "video", nerdfonts.Video, style.HeadlineVideoColor},
	{"[pdf]", "pdf", nerdfonts.Document, style.HeadlinePDFColor},
	{"[PDF]", "PDF", nerdfonts.Document, style.HeadlinePDFColor},
}

// parse tokenizes a plain title. An HN prefix counts only at the very start
// of the title — a quoted "Show HN:" mid-sentence stays prose.
func parse(title string) []span {
	spans := splitHNPrefix(title)
	spans = splitPlain(spans, func(text string) []span { return splitByRegexp(text, reYCLabel, spanYCLabel) })
	spans = splitPlain(spans, func(text string) []span { return splitByRegexp(text, reYear, spanYear) })
	spans = splitPlain(spans, splitContentTags)

	return spans
}

func splitHNPrefix(title string) []span {
	for _, p := range hnPrefixes {
		if rest, found := strings.CutPrefix(title, p.literal); found {
			out := []span{{text: p.literal, kind: spanHNPrefix}}
			if rest != "" {
				out = append(out, span{text: rest, kind: spanPlain})
			}

			return out
		}
	}

	return []span{{text: title, kind: spanPlain}}
}

// splitPlain rewrites each plain span through fn; token spans pass through
// untouched, so tokenizers cannot corrupt each other's output. fn returns
// nil to keep the span unchanged.
func splitPlain(spans []span, fn func(text string) []span) []span {
	out := make([]span, 0, len(spans))

	for _, s := range spans {
		if s.kind != spanPlain {
			out = append(out, s)

			continue
		}

		if parts := fn(s.text); parts != nil {
			out = append(out, parts...)
		} else {
			out = append(out, s)
		}
	}

	return out
}

// splitByRegexp splits text at re's matches; the token span keeps only the
// first capture group, dropping the surrounding parentheses. A text without
// matches returns nil so the caller keeps the original span.
func splitByRegexp(text string, re *regexp.Regexp, kind spanKind) []span {
	locs := re.FindAllStringSubmatchIndex(text, -1)
	if locs == nil {
		return nil
	}

	var out []span

	last := 0

	for _, loc := range locs {
		if loc[0] > last {
			out = append(out, span{text: text[last:loc[0]], kind: spanPlain})
		}

		out = append(out, span{text: text[loc[2]:loc[3]], kind: kind})
		last = loc[1]
	}

	if last < len(text) {
		out = append(out, span{text: text[last:], kind: spanPlain})
	}

	return out
}

// splitContentTags splits at [audio]/[video]/[pdf] literals; the span keeps
// the bare word. Only the bracketed source form matches — a title that
// happens to contain a nerd-font glyph stays plain text.
func splitContentTags(text string) []span {
	var out []span

	rest := text

	for {
		best, bestLit, bestWord := -1, "", ""

		for _, tag := range contentTags {
			if idx := strings.Index(rest, tag.literal); idx >= 0 && (best < 0 || idx < best) {
				best, bestLit, bestWord = idx, tag.literal, tag.word
			}
		}

		if best < 0 {
			break
		}

		if best > 0 {
			out = append(out, span{text: rest[:best], kind: spanPlain})
		}

		out = append(out, span{text: bestWord, kind: spanContentTag})
		rest = rest[best+len(bestLit):]
	}

	if out == nil {
		return nil
	}

	if rest != "" {
		out = append(out, span{text: rest, kind: spanPlain})
	}

	return out
}

// Render styles title for the given row state. Every span renders as
// self-contained lipgloss output — token colors compose with the state's
// base style, and nothing is left open at the end of the string.
func Render(title string, state HighlightType, enableNerdFonts bool) string {
	var b strings.Builder

	base := baseStyle(state)

	for _, sp := range parse(title) {
		b.WriteString(renderSpan(sp, base, state, enableNerdFonts))
	}

	return b.String()
}

// baseStyle is the single source of per-state title styling: list rows and
// pane headers render plain text with it, and token styles compose with it.
func baseStyle(state HighlightType) lipgloss.Style {
	s := lipgloss.NewStyle()

	switch state {
	case Selected:
		return s.Reverse(true)
	case OpenStory:
		// A muted highlight bar: the color scheme's bright black as
		// background, text in the scheme's default foreground. Reversing a
		// bright-black foreground would look the same on most schemes but
		// renders the row invisible where bright black equals the background
		// (Solarized Dark).
		return s.Background(lipgloss.BrightBlack)
	case MarkAsRead:
		return s.Italic(true).Faint(true)
	case AddToFavorites:
		return s.Foreground(lipgloss.Green).Reverse(true)
	case RemoveFromFavorites:
		return s.Foreground(lipgloss.Red).Reverse(true)
	case HeadlineInCommentSection:
		return s.Bold(true)
	case Unselected:
	}

	return s
}

func renderSpan(sp span, base lipgloss.Style, state HighlightType, enableNerdFonts bool) string {
	switch sp.kind {
	case spanHNPrefix:
		return base.Foreground(prefixColor(sp.text)).Render(sp.text)

	case spanYCLabel:
		if enableNerdFonts {
			return ycPill(sp.text, state)
		}

		return base.Foreground(style.HeadlineYCLabelColor()).Render(sp.text)

	case spanYear:
		return base.Foreground(style.HeadlineYearColor()).Render(sp.text)

	case spanContentTag:
		text, c := tagDisplay(sp.text, enableNerdFonts)

		return base.Foreground(c).Render(text)

	case spanPlain:
	}

	return base.Render(sp.text)
}

func prefixColor(literal string) color.Color {
	for _, p := range hnPrefixes {
		if p.literal == literal {
			return p.color()
		}
	}

	return lipgloss.NoColor{}
}

func tagDisplay(word string, enableNerdFonts bool) (string, color.Color) {
	for _, tag := range contentTags {
		if tag.word == word {
			if enableNerdFonts {
				return tag.glyph, tag.color()
			}

			return tag.word, tag.color()
		}
	}

	return word, lipgloss.NoColor{}
}

// ycPill renders the YC label as a nerd-font pill: the YCombinator glyph and
// the batch season on a colored bar between powerline separators. Selected
// swaps the pill's colors so it stays legible on the reversed row.
func ycPill(label string, state HighlightType) string {
	text := nerdfonts.YCombinator + noBreakSpace + strings.TrimPrefix(label, "YC ")

	c := style.HeadlineYCLabelColor()
	black := lipgloss.ANSIColor(ansiBlack)

	var fg, bg color.Color = black, c
	if state == Selected {
		fg, bg = c, black
	}

	content := lipgloss.NewStyle().Foreground(fg).Background(bg)

	switch state {
	case MarkAsRead:
		content = content.Italic(true).Faint(true)
	case HeadlineInCommentSection:
		content = content.Bold(true)
	case Unselected, Selected, OpenStory, AddToFavorites, RemoveFromFavorites:
	}

	border := pillBorder(bg, state)

	return border.Render(nerdfonts.LeftSeparator) + content.Render(text) + border.Render(nerdfonts.RightSeparator)
}

// pillBorder styles the powerline separators framing the pill so they blend
// into the row behind them: reversed on the Selected bar, on the OpenStory
// background elsewhere.
func pillBorder(bg color.Color, state HighlightType) lipgloss.Style {
	switch state {
	case Selected:
		return lipgloss.NewStyle().Foreground(lipgloss.NoColor{}).Background(bg).Reverse(true)
	case OpenStory:
		return lipgloss.NewStyle().Foreground(bg).Background(lipgloss.BrightBlack)
	case Unselected, HeadlineInCommentSection, MarkAsRead, AddToFavorites, RemoveFromFavorites:
	}

	return lipgloss.NewStyle().Foreground(bg)
}

func HighlightDomain(domain string) string {
	if domain == "" {
		return ""
	}

	return style.Faint("(" + domain + ")")
}
