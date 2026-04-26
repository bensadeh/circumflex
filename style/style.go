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

// Static styles — built once from lipgloss-builtin colors.
var (
	redStyle          = lipgloss.NewStyle().Foreground(lipgloss.Red)
	blueStyle         = lipgloss.NewStyle().Foreground(lipgloss.Blue)
	greenStyle        = lipgloss.NewStyle().Foreground(lipgloss.Green)
	yellowStyle       = lipgloss.NewStyle().Foreground(lipgloss.Yellow)
	magentaStyle      = lipgloss.NewStyle().Foreground(lipgloss.Magenta)
	cyanStyle         = lipgloss.NewStyle().Foreground(lipgloss.Cyan)
	whiteStyle        = lipgloss.NewStyle().Foreground(lipgloss.White)
	brightRedStyle    = lipgloss.NewStyle().Foreground(lipgloss.BrightRed)
	brightGreenStyle  = lipgloss.NewStyle().Foreground(lipgloss.BrightGreen)
	brightYellowStyle = lipgloss.NewStyle().Foreground(lipgloss.BrightYellow)
	brightWhiteStyle  = lipgloss.NewStyle().Foreground(lipgloss.BrightWhite)
	boldStyle         = lipgloss.NewStyle().Bold(true)
	boldReverseStyle  = lipgloss.NewStyle().Bold(true).Reverse(true)
	faintStyle        = lipgloss.NewStyle().Faint(true)
	faintItalicStyle  = lipgloss.NewStyle().Faint(true).Italic(true)
)

// Theme-dependent styles — rebuilt by rebuildThemeStyles whenever the theme changes.
var (
	headlineAskHNStyle    lipgloss.Style
	headlineShowHNStyle   lipgloss.Style
	headlineTellHNStyle   lipgloss.Style
	headlineThankHNStyle  lipgloss.Style
	headlineLaunchHNStyle lipgloss.Style
	headlineAudioStyle    lipgloss.Style
	headlineVideoStyle    lipgloss.Style
	headlinePDFStyle      lipgloss.Style

	commentURLStyle          lipgloss.Style // hyperlink applied at call time
	commentMentionStyle      lipgloss.Style
	commentModStyle          lipgloss.Style
	commentVariableStyle     lipgloss.Style
	commentOPStyle           lipgloss.Style
	commentGPStyle           lipgloss.Style
	commentNewIndicatorStyle lipgloss.Style
	commentBacktickStyle     lipgloss.Style

	metaAuthorStyle      lipgloss.Style
	metaCommentsStyle    lipgloss.Style
	metaScoreStyle       lipgloss.Style
	metaNewCommentsStyle lipgloss.Style
	metaURLStyle         lipgloss.Style // hyperlink applied at call time
	metaReaderModeStyle  lipgloss.Style

	readerH1Style lipgloss.Style
	readerH2Style lipgloss.Style
	readerH3Style lipgloss.Style
	readerH4Style lipgloss.Style
	readerH5Style lipgloss.Style
	readerH6Style lipgloss.Style

	logoCStyle lipgloss.Style
	logoLStyle lipgloss.Style
	logoXStyle lipgloss.Style
)

// Theme-dependent colors — rebuilt by rebuildThemeStyles whenever the theme changes.
var (
	headlineYCLabelColor color.Color
	headlineYearColor    color.Color

	commentModColor color.Color

	metaIDColor color.Color

	readerBBCImageColor   color.Color
	readerBBCCaptionColor color.Color
	readerImageColor      color.Color

	headerCColor         color.Color
	headerLColor         color.Color
	headerXColor         color.Color
	headerFavoritesColor color.Color
	headerPrimaryColor   color.Color
	headerSecondaryColor color.Color
	headerTertiaryColor  color.Color
)

var indentCycleFuncs []func(string) string

//nolint:gochecknoinits // populate theme-dependent styles with defaults before any caller (including tests) uses them
func init() {
	rebuildThemeStyles()
}

func Init(t *theme.Theme) {
	current = t

	rebuildThemeStyles()
}

func rebuildThemeStyles() {
	fg := func(c string) lipgloss.Style {
		return lipgloss.NewStyle().Foreground(theme.ParseColor(c))
	}

	headlineAskHNStyle = fg(current.Headline.AskHN)
	headlineShowHNStyle = fg(current.Headline.ShowHN)
	headlineTellHNStyle = fg(current.Headline.TellHN)
	headlineThankHNStyle = fg(current.Headline.ThankHN)
	headlineLaunchHNStyle = fg(current.Headline.LaunchHN)
	headlineAudioStyle = fg(current.Headline.Audio)
	headlineVideoStyle = fg(current.Headline.Video)
	headlinePDFStyle = fg(current.Headline.PDF)
	headlineYCLabelColor = theme.ParseColor(current.Headline.YCLabel)
	headlineYearColor = theme.ParseColor(current.Headline.Year)

	commentURLStyle = fg(current.Comment.URL)
	commentMentionStyle = fg(current.Comment.Mention)
	commentModStyle = fg(current.Comment.Mod)
	commentVariableStyle = fg(current.Comment.Variable)
	commentOPStyle = fg(current.Comment.OP)
	commentGPStyle = fg(current.Comment.GP)
	commentNewIndicatorStyle = fg(current.Comment.NewIndicator)
	commentBacktickStyle = fg(current.Comment.Backtick).Italic(true)
	commentModColor = theme.ParseColor(current.Comment.Mod)

	metaAuthorStyle = fg(current.Meta.Author)
	metaCommentsStyle = fg(current.Meta.Comments)
	metaScoreStyle = fg(current.Meta.Score)
	metaNewCommentsStyle = fg(current.Meta.NewComments)
	metaURLStyle = fg(current.Meta.URL)
	metaReaderModeStyle = fg(current.Meta.ReaderMode)
	metaIDColor = theme.ParseColor(current.Meta.ID)

	readerH1Style = fg(current.Reader.H1)
	readerH2Style = fg(current.Reader.H2)
	readerH3Style = fg(current.Reader.H3)
	readerH4Style = fg(current.Reader.H4)
	readerH5Style = fg(current.Reader.H5)
	readerH6Style = fg(current.Reader.H6)
	readerBBCImageColor = theme.ParseColor(current.Reader.BBCImage)
	readerBBCCaptionColor = theme.ParseColor(current.Reader.BBCCaption)
	readerImageColor = theme.ParseColor(current.Reader.Image)

	headerCColor = theme.ParseColor(current.Header.C)
	headerLColor = theme.ParseColor(current.Header.L)
	headerXColor = theme.ParseColor(current.Header.X)
	headerFavoritesColor = theme.ParseColor(current.Header.Favorites)
	headerPrimaryColor = theme.ParseColor(current.App.Primary)
	headerSecondaryColor = theme.ParseColor(current.App.Secondary)
	headerTertiaryColor = theme.ParseColor(current.App.Tertiary)

	logoCStyle = lipgloss.NewStyle().Foreground(headerCColor)
	logoLStyle = lipgloss.NewStyle().Foreground(headerLColor)
	logoXStyle = lipgloss.NewStyle().Foreground(headerXColor)

	indentCycleFuncs = make([]func(string) string, len(current.Indent.Cycle))
	for i, c := range current.Indent.Cycle {
		s := fg(c)
		indentCycleFuncs[i] = func(text string) string { return render(s, text) }
	}
}

// SetBaseForeground sets the base foreground color that render() will
// re-apply after each styled span's reset. Call ClearBaseForeground when done.
func SetBaseForeground(c color.Color) {
	baseFg = foregroundCode(c)
}

func ClearBaseForeground() {
	baseFg = ""
}

func CommentModFg() color.Color { return commentModColor }

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

func Red(s string) string          { return render(redStyle, s) }
func Blue(s string) string         { return render(blueStyle, s) }
func Green(s string) string        { return render(greenStyle, s) }
func Yellow(s string) string       { return render(yellowStyle, s) }
func Magenta(s string) string      { return render(magentaStyle, s) }
func Cyan(s string) string         { return render(cyanStyle, s) }
func White(s string) string        { return render(whiteStyle, s) }
func BrightRed(s string) string    { return render(brightRedStyle, s) }
func BrightGreen(s string) string  { return render(brightGreenStyle, s) }
func BrightYellow(s string) string { return render(brightYellowStyle, s) }
func BrightWhite(s string) string  { return render(brightWhiteStyle, s) }

func Bold(s string) string        { return boldStyle.Render(s) }
func BoldReverse(s string) string { return boldReverseStyle.Render(s) }
func Faint(s string) string       { return faintStyle.Render(s) }
func FaintItalic(s string) string { return faintItalicStyle.Render(s) }

func HeadlineAskHN(s string) string    { return render(headlineAskHNStyle, s) }
func HeadlineShowHN(s string) string   { return render(headlineShowHNStyle, s) }
func HeadlineTellHN(s string) string   { return render(headlineTellHNStyle, s) }
func HeadlineThankHN(s string) string  { return render(headlineThankHNStyle, s) }
func HeadlineLaunchHN(s string) string { return render(headlineLaunchHNStyle, s) }
func HeadlineAudio(s string) string    { return render(headlineAudioStyle, s) }
func HeadlineVideo(s string) string    { return render(headlineVideoStyle, s) }
func HeadlinePDF(s string) string      { return render(headlinePDFStyle, s) }

func HeadlineYCLabelColor() color.Color { return headlineYCLabelColor }
func HeadlineYearColor() color.Color    { return headlineYearColor }

func CommentURL(s, url string) string     { return render(commentURLStyle.Hyperlink(url), s) }
func CommentMention(s string) string      { return render(commentMentionStyle, s) }
func CommentMod(s string) string          { return render(commentModStyle, s) }
func CommentVariable(s string) string     { return render(commentVariableStyle, s) }
func CommentOP(s string) string           { return render(commentOPStyle, s) }
func CommentGP(s string) string           { return render(commentGPStyle, s) }
func CommentNewIndicator(s string) string { return render(commentNewIndicatorStyle, s) }
func CommentBacktick(s string) string     { return render(commentBacktickStyle, s) }

func MetaAuthor(s string) string      { return render(metaAuthorStyle, s) }
func MetaComments(s string) string    { return render(metaCommentsStyle, s) }
func MetaScore(s string) string       { return render(metaScoreStyle, s) }
func MetaNewComments(s string) string { return render(metaNewCommentsStyle, s) }
func MetaURL(s, url string) string    { return render(metaURLStyle.Hyperlink(url), s) }
func MetaReaderMode(s string) string  { return render(metaReaderModeStyle, s) }
func MetaIDColor() color.Color        { return metaIDColor }

func ReaderH1(s string) string           { return render(readerH1Style, s) }
func ReaderH2(s string) string           { return render(readerH2Style, s) }
func ReaderH3(s string) string           { return render(readerH3Style, s) }
func ReaderH4(s string) string           { return render(readerH4Style, s) }
func ReaderH5(s string) string           { return render(readerH5Style, s) }
func ReaderH6(s string) string           { return render(readerH6Style, s) }
func ReaderBBCImageColor() color.Color   { return readerBBCImageColor }
func ReaderBBCCaptionColor() color.Color { return readerBBCCaptionColor }
func ReaderImageColor() color.Color      { return readerImageColor }

func HeaderC() color.Color         { return headerCColor }
func HeaderL() color.Color         { return headerLColor }
func HeaderX() color.Color         { return headerXColor }
func HeaderFavorites() color.Color { return headerFavoritesColor }
func HeaderPrimary() color.Color   { return headerPrimaryColor }
func HeaderSecondary() color.Color { return headerSecondaryColor }
func HeaderTertiary() color.Color  { return headerTertiaryColor }

func Logo(a, b, c string) string {
	return logoCStyle.Render(a) + logoLStyle.Render(b) + logoXStyle.Render(c)
}

func IndentCycle() []func(string) string { return indentCycleFuncs }

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
