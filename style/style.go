package style

import (
	"image/color"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/theme"

	"charm.land/lipgloss/v2"
)

const IndentSymbol = "▎"

// PrefixLines prepends prefix to every line in s.
// A trailing empty line (from a final \n) is kept but not prefixed,
// so that concatenating two prefixed blocks doesn't double-indent.
func PrefixLines(s, prefix string) string {
	lines := strings.Split(s, "\n")
	last := len(lines) - 1

	for i, line := range lines {
		if i == last && line == "" {
			break
		}

		lines[i] = prefix + line
	}

	return strings.Join(lines, "\n")
}

var current = theme.Default()

// baseFg holds a raw ANSI foreground escape to re-apply after every
// styled span's reset. Set before rendering a mod paragraph, cleared after.
// Safe because the Bubble Tea rendering pipeline is single-threaded.
var baseFg string

func Init(t *theme.Theme) {
	current = t
}

// SetBaseForeground sets the base foreground color that render() will
// re-apply after each styled span's reset. Call ClearBaseForeground when done.
func SetBaseForeground(c color.Color) {
	baseFg = foregroundCode(c)
}

func ClearBaseForeground() {
	baseFg = ""
}

func CommentModFg() color.Color { return theme.ParseColor(current.Comment.Mod) }

// render is the single chokepoint for all lipgloss foreground renders.
// It appends baseFg (when set) so the mod tint automatically resumes
// after every styled span's reset.
func render(s lipgloss.Style, text string) string {
	result := s.Render(text)
	if baseFg != "" {
		result += baseFg
	}

	return result
}

func colored(colorStr, text string) string {
	return render(lipgloss.NewStyle().Foreground(theme.ParseColor(colorStr)), text)
}

func coloredLinked(colorStr, text, url string) string {
	return render(lipgloss.NewStyle().Foreground(theme.ParseColor(colorStr)).Hyperlink(url), text)
}

func Red(s string) string       { return render(lipgloss.NewStyle().Foreground(lipgloss.Red), s) }
func Blue(s string) string      { return render(lipgloss.NewStyle().Foreground(lipgloss.Blue), s) }
func Green(s string) string     { return render(lipgloss.NewStyle().Foreground(lipgloss.Green), s) }
func Yellow(s string) string    { return render(lipgloss.NewStyle().Foreground(lipgloss.Yellow), s) }
func Magenta(s string) string   { return render(lipgloss.NewStyle().Foreground(lipgloss.Magenta), s) }
func Cyan(s string) string      { return render(lipgloss.NewStyle().Foreground(lipgloss.Cyan), s) }
func White(s string) string     { return render(lipgloss.NewStyle().Foreground(lipgloss.White), s) }
func BrightRed(s string) string { return render(lipgloss.NewStyle().Foreground(lipgloss.BrightRed), s) }
func BrightGreen(s string) string {
	return render(lipgloss.NewStyle().Foreground(lipgloss.BrightGreen), s)
}

func BrightYellow(s string) string {
	return render(lipgloss.NewStyle().Foreground(lipgloss.BrightYellow), s)
}

func BrightWhite(s string) string {
	return render(lipgloss.NewStyle().Foreground(lipgloss.BrightWhite), s)
}

func Bold(s string) string {
	return lipgloss.NewStyle().Bold(true).Render(s)
}

func BoldReverse(s string) string {
	return lipgloss.NewStyle().Bold(true).Reverse(true).Render(s)
}

func Faint(s string) string {
	return lipgloss.NewStyle().Faint(true).Render(s)
}

func FaintItalic(s string) string {
	return lipgloss.NewStyle().Faint(true).Italic(true).Render(s)
}

func HeadlineAskHN(s string) string    { return colored(current.Headline.AskHN, s) }
func HeadlineShowHN(s string) string   { return colored(current.Headline.ShowHN, s) }
func HeadlineTellHN(s string) string   { return colored(current.Headline.TellHN, s) }
func HeadlineThankHN(s string) string  { return colored(current.Headline.ThankHN, s) }
func HeadlineLaunchHN(s string) string { return colored(current.Headline.LaunchHN, s) }
func HeadlineAudio(s string) string    { return colored(current.Headline.Audio, s) }
func HeadlineVideo(s string) string    { return colored(current.Headline.Video, s) }
func HeadlinePDF(s string) string      { return colored(current.Headline.PDF, s) }

func HeadlineYCLabelColor() color.Color { return theme.ParseColor(current.Headline.YCLabel) }
func HeadlineYearColor() color.Color    { return theme.ParseColor(current.Headline.Year) }

func CommentURL(s, url string) string     { return coloredLinked(current.Comment.URL, s, url) }
func CommentMention(s string) string      { return colored(current.Comment.Mention, s) }
func CommentMod(s string) string          { return colored(current.Comment.Mod, s) }
func CommentVariable(s string) string     { return colored(current.Comment.Variable, s) }
func CommentOP(s string) string           { return colored(current.Comment.OP, s) }
func CommentGP(s string) string           { return colored(current.Comment.GP, s) }
func CommentNewIndicator(s string) string { return colored(current.Comment.NewIndicator, s) }

func CommentBacktick(s string) string {
	return render(lipgloss.NewStyle().Foreground(theme.ParseColor(current.Comment.Backtick)).Italic(true), s)
}

func MetaAuthor(s string) string      { return colored(current.Meta.Author, s) }
func MetaComments(s string) string    { return colored(current.Meta.Comments, s) }
func MetaScore(s string) string       { return colored(current.Meta.Score, s) }
func MetaNewComments(s string) string { return colored(current.Meta.NewComments, s) }
func MetaURL(s, url string) string    { return coloredLinked(current.Meta.URL, s, url) }
func MetaReaderMode(s string) string  { return colored(current.Meta.ReaderMode, s) }
func MetaIDColor() color.Color        { return theme.ParseColor(current.Meta.ID) }

func ReaderH1(s string) string           { return colored(current.Reader.H1, s) }
func ReaderH2(s string) string           { return colored(current.Reader.H2, s) }
func ReaderH3(s string) string           { return colored(current.Reader.H3, s) }
func ReaderH4(s string) string           { return colored(current.Reader.H4, s) }
func ReaderH5(s string) string           { return colored(current.Reader.H5, s) }
func ReaderH6(s string) string           { return colored(current.Reader.H6, s) }
func ReaderBBCImageColor() color.Color   { return theme.ParseColor(current.Reader.BBCImage) }
func ReaderBBCCaptionColor() color.Color { return theme.ParseColor(current.Reader.BBCCaption) }
func ReaderImageColor() color.Color      { return theme.ParseColor(current.Reader.Image) }

func HeaderC() color.Color         { return theme.ParseColor(current.Header.C) }
func HeaderL() color.Color         { return theme.ParseColor(current.Header.L) }
func HeaderX() color.Color         { return theme.ParseColor(current.Header.X) }
func HeaderFavorites() color.Color { return theme.ParseColor(current.Header.Favorites) }
func HeaderPrimary() color.Color   { return theme.ParseColor(current.App.Primary) }
func HeaderSecondary() color.Color { return theme.ParseColor(current.App.Secondary) }
func HeaderTertiary() color.Color  { return theme.ParseColor(current.App.Tertiary) }

func Logo(a, b, c string) string {
	cs := lipgloss.NewStyle().Foreground(HeaderC())
	ls := lipgloss.NewStyle().Foreground(HeaderL())
	xs := lipgloss.NewStyle().Foreground(HeaderX())

	return cs.Render(a) + ls.Render(b) + xs.Render(c)
}

func IndentCycle() []func(string) string {
	funcs := make([]func(string) string, len(current.Indent.Cycle))
	for i, c := range current.Indent.Cycle {
		colorStr := c
		funcs[i] = func(s string) string { return colored(colorStr, s) }
	}

	return funcs
}

type Binding struct {
	Key  string
	Desc string
}

func ModeIndicator(bindings []Binding) string {
	parts := make([]string, len(bindings))
	for i, b := range bindings {
		parts[i] = RenderBinding(b)
	}

	return strings.Repeat(" ", 2) + strings.Join(parts, "  ")
}

func RenderBinding(b Binding) string {
	return b.Key + Faint(": "+b.Desc)
}

// PaintForeground tints text with the given foreground color. Lipgloss resets
// are already handled by render()/baseFg; this covers direct ansi.Reset
// usage (e.g. </i> tag replacement) and per-line prefixing so the color
// survives external indent-symbol prefixes.
func PaintForeground(s string, c color.Color) string {
	prefix := foregroundCode(c)
	if prefix == "" {
		return s
	}

	// Re-apply color after our own ansi.Reset sequences
	// (injected by syntax.ReplaceHTML for closing </i> tags).
	s = strings.ReplaceAll(s, ansi.Reset, ansi.Reset+prefix)

	// Prepend color to every non-empty line so it survives external
	// prefixing (indent symbols, margins) that includes its own resets.
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}

	return strings.Join(lines, "\n") + ansi.Reset
}

// foregroundCode returns the raw ANSI foreground escape for a color.Color,
// with no trailing reset.
func foregroundCode(c color.Color) string {
	const marker = "\xff"

	rendered := lipgloss.NewStyle().Foreground(c).Render(marker)
	idx := strings.Index(rendered, marker)

	if idx <= 0 {
		return ""
	}

	return rendered[:idx]
}
