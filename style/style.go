package style

import (
	"clx/theme"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

var current = theme.Default()

func Init(t *theme.Theme) {
	current = t
}

func colored(colorStr, text string) string {
	return lipgloss.NewStyle().Foreground(theme.ParseColor(colorStr)).Render(text)
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

func BrightBlue(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.BrightBlue).Render(s)
}

func BrightGreen(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.BrightGreen).Render(s)
}

func BrightYellow(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.BrightYellow).Render(s)
}

func BrightCyan(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.BrightCyan).Render(s)
}

func BrightMagenta(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.BrightMagenta).Render(s)
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

func CommentURL(s string) string          { return colored(current.Comment.URL, s) }
func CommentMention(s string) string      { return colored(current.Comment.Mention, s) }
func CommentMod(s string) string          { return colored(current.Comment.Mod, s) }
func CommentVariable(s string) string     { return colored(current.Comment.Variable, s) }
func CommentOP(s string) string           { return colored(current.Comment.OP, s) }
func CommentGP(s string) string           { return colored(current.Comment.GP, s) }
func CommentNewIndicator(s string) string { return colored(current.Comment.NewIndicator, s) }

func CommentBacktickColor() string {
	return ansiColorCode(current.Comment.Backtick)
}

// Meta colors.

func MetaAuthor(s string) string      { return colored(current.Meta.Author, s) }
func MetaComments(s string) string    { return colored(current.Meta.Comments, s) }
func MetaScore(s string) string       { return colored(current.Meta.Score, s) }
func MetaNewComments(s string) string { return colored(current.Meta.NewComments, s) }
func MetaURL(s string) string         { return colored(current.Meta.URL, s) }
func MetaReaderMode(s string) string  { return colored(current.Meta.ReaderMode, s) }
func MetaIDColor() color.Color        { return theme.ParseColor(current.Meta.ID) }

// Reader colors.

func ReaderH1(s string) string           { return colored(current.Reader.H1, s) }
func ReaderH2(s string) string           { return colored(current.Reader.H2, s) }
func ReaderH3(s string) string           { return colored(current.Reader.H3, s) }
func ReaderH4(s string) string           { return colored(current.Reader.H4, s) }
func ReaderH5(s string) string           { return colored(current.Reader.H5, s) }
func ReaderH6(s string) string           { return colored(current.Reader.H6, s) }
func ReaderBBCImage(s string) string     { return colored(current.Reader.BBCImage, s) }
func ReaderBBCCaption(s string) string   { return colored(current.Reader.BBCCaption, s) }
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

// Help colors.

func HelpMainMenu(s string) string       { return colored(current.Help.MainMenu, s) }
func HelpCommentSection(s string) string { return colored(current.Help.CommentSection, s) }
func HelpLegend(s string) string         { return colored(current.Help.Legend, s) }
func HelpMainMenuColor() color.Color     { return theme.ParseColor(current.Help.MainMenu) }
func HelpCommentColor() color.Color      { return theme.ParseColor(current.Help.CommentSection) }
func HelpLegendColor() color.Color       { return theme.ParseColor(current.Help.Legend) }

// Indent colors.

func IndentCycle() []func(string) string {
	funcs := make([]func(string) string, len(current.Indent.Cycle))
	for i, c := range current.Indent.Cycle {
		colorStr := c
		funcs[i] = func(s string) string { return colored(colorStr, s) }
	}

	return funcs
}

// Footer colors.

func FooterReadMode() color.Color     { return theme.ParseColor(current.Footer.ReadMode) }
func FooterNavigateMode() color.Color { return theme.ParseColor(current.Footer.NavigateMode) }

const modeWidth = 9

type Binding struct {
	Key  string
	Desc string
}

func ModeIndicator(mode string, modeColor color.Color, leftMargin, screenWidth int, logo string, bindings []Binding) string {
	modeStyle := lipgloss.NewStyle().
		Foreground(modeColor).
		Width(modeWidth).
		Align(lipgloss.Center).
		MarginLeft(leftMargin)

	parts := make([]string, len(bindings))
	for i, b := range bindings {
		parts[i] = b.Key + Faint(": "+b.Desc)
	}

	left := modeStyle.Render(mode) + Faint("│ ") + strings.Join(parts, "  ")
	leftWidth := lipgloss.Width(left)
	logoWidth := lipgloss.Width(logo)
	filler := strings.Repeat(" ", max(0, screenWidth-leftWidth-logoWidth-1))

	return left + filler + logo + " "
}

// Helpers.

func ansiColorCode(colorName string) string {
	codes := map[string]string{
		"red":            "\033[31m",
		"green":          "\033[32m",
		"yellow":         "\033[33m",
		"blue":           "\033[34m",
		"magenta":        "\033[35m",
		"cyan":           "\033[36m",
		"white":          "\033[37m",
		"bright_red":     "\033[91m",
		"bright_green":   "\033[92m",
		"bright_yellow":  "\033[93m",
		"bright_blue":    "\033[94m",
		"bright_magenta": "\033[95m",
		"bright_cyan":    "\033[96m",
		"bright_white":   "\033[97m",
	}

	if code, ok := codes[colorName]; ok {
		return code
	}

	return "\033[35m"
}
