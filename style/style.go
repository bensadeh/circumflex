package style

import (
	"image/color"
	"strings"

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

func Init(t *theme.Theme) {
	current = t
}

func colored(colorStr, text string) string {
	return lipgloss.NewStyle().Foreground(theme.ParseColor(colorStr)).Render(text)
}

func coloredLinked(colorStr, text, url string) string {
	return lipgloss.NewStyle().Foreground(theme.ParseColor(colorStr)).Hyperlink(url).Render(text)
}

// Generic color helpers (for CLI messages, non-themed uses).

func Red(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Red).Render(s)
}

func Blue(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Blue).Render(s)
}

func Green(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Green).Render(s)
}

func Yellow(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Yellow).Render(s)
}

func Magenta(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Magenta).Render(s)
}

func Cyan(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Cyan).Render(s)
}

func White(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.White).Render(s)
}

func BrightRed(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.BrightRed).Render(s)
}

func BrightGreen(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.BrightGreen).Render(s)
}

func BrightYellow(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.BrightYellow).Render(s)
}

func BrightWhite(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.BrightWhite).Render(s)
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

// Semantic helpers — theme-aware.

// Headline colors.

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

// Comment colors.

func CommentURL(s, url string) string     { return coloredLinked(current.Comment.URL, s, url) }
func CommentMention(s string) string      { return colored(current.Comment.Mention, s) }
func CommentMod(s string) string          { return colored(current.Comment.Mod, s) }
func CommentVariable(s string) string     { return colored(current.Comment.Variable, s) }
func CommentOP(s string) string           { return colored(current.Comment.OP, s) }
func CommentGP(s string) string           { return colored(current.Comment.GP, s) }
func CommentNewIndicator(s string) string { return colored(current.Comment.NewIndicator, s) }

func CommentBacktick(s string) string {
	return lipgloss.NewStyle().Foreground(theme.ParseColor(current.Comment.Backtick)).Italic(true).Render(s)
}

// Meta colors.

func MetaAuthor(s string) string      { return colored(current.Meta.Author, s) }
func MetaComments(s string) string    { return colored(current.Meta.Comments, s) }
func MetaScore(s string) string       { return colored(current.Meta.Score, s) }
func MetaNewComments(s string) string { return colored(current.Meta.NewComments, s) }
func MetaURL(s, url string) string    { return coloredLinked(current.Meta.URL, s, url) }
func MetaReaderMode(s string) string  { return colored(current.Meta.ReaderMode, s) }
func MetaIDColor() color.Color        { return theme.ParseColor(current.Meta.ID) }

// Reader colors.

func ReaderH1(s string) string           { return colored(current.Reader.H1, s) }
func ReaderH2(s string) string           { return colored(current.Reader.H2, s) }
func ReaderH3(s string) string           { return colored(current.Reader.H3, s) }
func ReaderH4(s string) string           { return colored(current.Reader.H4, s) }
func ReaderH5(s string) string           { return colored(current.Reader.H5, s) }
func ReaderH6(s string) string           { return colored(current.Reader.H6, s) }
func ReaderBBCImageColor() color.Color   { return theme.ParseColor(current.Reader.BBCImage) }
func ReaderBBCCaptionColor() color.Color { return theme.ParseColor(current.Reader.BBCCaption) }
func ReaderImageColor() color.Color      { return theme.ParseColor(current.Reader.Image) }

// Header colors.

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

// Indent colors.

func IndentCycle() []func(string) string {
	funcs := make([]func(string) string, len(current.Indent.Cycle))
	for i, c := range current.Indent.Cycle {
		colorStr := c
		funcs[i] = func(s string) string { return colored(colorStr, s) }
	}

	return funcs
}

// Footer.

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
