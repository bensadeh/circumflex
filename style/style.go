package style

import (
	"image/color"
	"regexp"
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

	memorialColor          = lipgloss.ANSIColor(8)
	memorialUnderlineStyle = lipgloss.NewStyle().Foreground(memorialColor)
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

	logoCFaintStyle lipgloss.Style
	logoLFaintStyle lipgloss.Style
	logoXFaintStyle lipgloss.Style
)

// Theme-dependent colors — rebuilt by rebuildThemeStyles whenever the theme changes.
var (
	headlineYCLabelColor color.Color
	headlineYearColor    color.Color

	commentModColor color.Color

	metaIDColor color.Color

	readerImageColor color.Color

	headerCColor         color.Color
	headerLColor         color.Color
	headerXColor         color.Color
	headerPrimaryColor   color.Color
	headerSecondaryColor color.Color
	headerTertiaryColor  color.Color
)

var indentCycleFuncs []func(string) string

//nolint:gochecknoinits // populate theme-dependent styles with defaults before any caller (including tests) uses them
func init() {
	rebuildThemeStyles()
}

// SetTheme swaps the package-global theme every render helper reads. It is a
// deliberate global: one process, one theme. Commands call it once at startup,
// before any rendering; it is not safe to call concurrently with rendering.
func SetTheme(t *theme.Theme) {
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
	readerImageColor = theme.ParseColor(current.Reader.Image)

	headerCColor = theme.ParseColor(current.Header.C)
	headerLColor = theme.ParseColor(current.Header.L)
	headerXColor = theme.ParseColor(current.Header.X)
	headerPrimaryColor = theme.ParseColor(current.App.Primary)
	headerSecondaryColor = theme.ParseColor(current.App.Secondary)
	headerTertiaryColor = theme.ParseColor(current.App.Tertiary)

	logoCStyle = lipgloss.NewStyle().Foreground(headerCColor)
	logoLStyle = lipgloss.NewStyle().Foreground(headerLColor)
	logoXStyle = lipgloss.NewStyle().Foreground(headerXColor)
	logoCFaintStyle = logoCStyle.Faint(true)
	logoLFaintStyle = logoLStyle.Faint(true)
	logoXFaintStyle = logoXStyle.Faint(true)

	indentCycleFuncs = make([]func(string) string, len(current.Indent.Cycle))
	for i, c := range current.Indent.Cycle {
		s := fg(c)
		indentCycleFuncs[i] = func(text string) string { return s.Render(text) }
	}
}

func CommentModFg() color.Color { return commentModColor }

func Red(s string) string          { return redStyle.Render(s) }
func Blue(s string) string         { return blueStyle.Render(s) }
func Green(s string) string        { return greenStyle.Render(s) }
func Yellow(s string) string       { return yellowStyle.Render(s) }
func Magenta(s string) string      { return magentaStyle.Render(s) }
func Cyan(s string) string         { return cyanStyle.Render(s) }
func White(s string) string        { return whiteStyle.Render(s) }
func BrightRed(s string) string    { return brightRedStyle.Render(s) }
func BrightGreen(s string) string  { return brightGreenStyle.Render(s) }
func BrightYellow(s string) string { return brightYellowStyle.Render(s) }
func BrightWhite(s string) string  { return brightWhiteStyle.Render(s) }

func Bold(s string) string              { return boldStyle.Render(s) }
func BoldReverse(s string) string       { return boldReverseStyle.Render(s) }
func Faint(s string) string             { return faintStyle.Render(s) }
func MemorialUnderline(s string) string { return memorialUnderlineStyle.Render(s) }
func MemorialColor() color.Color        { return memorialColor }
func FaintItalic(s string) string       { return faintItalicStyle.Render(s) }

func HeadlineAskHN(s string) string    { return headlineAskHNStyle.Render(s) }
func HeadlineShowHN(s string) string   { return headlineShowHNStyle.Render(s) }
func HeadlineTellHN(s string) string   { return headlineTellHNStyle.Render(s) }
func HeadlineThankHN(s string) string  { return headlineThankHNStyle.Render(s) }
func HeadlineLaunchHN(s string) string { return headlineLaunchHNStyle.Render(s) }
func HeadlineAudio(s string) string    { return headlineAudioStyle.Render(s) }
func HeadlineVideo(s string) string    { return headlineVideoStyle.Render(s) }
func HeadlinePDF(s string) string      { return headlinePDFStyle.Render(s) }

func HeadlineYCLabelColor() color.Color { return headlineYCLabelColor }
func HeadlineYearColor() color.Color    { return headlineYearColor }

func CommentURL(s, url string) string     { return commentURLStyle.Hyperlink(url).Render(s) }
func CommentMention(s string) string      { return commentMentionStyle.Render(s) }
func CommentMod(s string) string          { return commentModStyle.Render(s) }
func CommentVariable(s string) string     { return commentVariableStyle.Render(s) }
func CommentOP(s string) string           { return commentOPStyle.Render(s) }
func CommentGP(s string) string           { return commentGPStyle.Render(s) }
func CommentNewIndicator(s string) string { return commentNewIndicatorStyle.Render(s) }
func CommentBacktick(s string) string     { return commentBacktickStyle.Render(s) }

func MetaAuthor(s string) string      { return metaAuthorStyle.Render(s) }
func MetaComments(s string) string    { return metaCommentsStyle.Render(s) }
func MetaScore(s string) string       { return metaScoreStyle.Render(s) }
func MetaNewComments(s string) string { return metaNewCommentsStyle.Render(s) }
func MetaURL(s, url string) string    { return metaURLStyle.Hyperlink(url).Render(s) }
func MetaReaderMode(s string) string  { return metaReaderModeStyle.Render(s) }
func MetaIDColor() color.Color        { return metaIDColor }

func ReaderH1(s string) string      { return readerH1Style.Render(s) }
func ReaderH2(s string) string      { return readerH2Style.Render(s) }
func ReaderH3(s string) string      { return readerH3Style.Render(s) }
func ReaderH4(s string) string      { return readerH4Style.Render(s) }
func ReaderH5(s string) string      { return readerH5Style.Render(s) }
func ReaderH6(s string) string      { return readerH6Style.Render(s) }
func ReaderImageColor() color.Color { return readerImageColor }

func HeaderC() color.Color         { return headerCColor }
func HeaderL() color.Color         { return headerLColor }
func HeaderX() color.Color         { return headerXColor }
func HeaderPrimary() color.Color   { return headerPrimaryColor }
func HeaderSecondary() color.Color { return headerSecondaryColor }
func HeaderTertiary() color.Color  { return headerTertiaryColor }

func Logo(a, b, c string) string {
	return logoCStyle.Render(a) + logoLStyle.Render(b) + logoXStyle.Render(c)
}

func LogoFaint(a, b, c string) string {
	return logoCFaintStyle.Render(a) + logoLFaintStyle.Render(b) + logoXFaintStyle.Render(c)
}

func IndentCycle() []func(string) string { return indentCycleFuncs }

// sgrResetPattern matches the two SGR-reset forms that appear in our rendered
// output: lipgloss emits the short form \x1b[m after styled spans, while our
// own ansi.Reset (used by syntax.ReplaceHTML for </i> etc.) is the long form
// \x1b[0m. Both wipe the foreground and so both need the tint reapplied.
var sgrResetPattern = regexp.MustCompile(`\x1b\[0?m`)

// PaintForeground re-applies the given foreground color throughout s so the
// tint survives every SGR reset (from lipgloss spans and our own ansi.Reset)
// and the outer indent-symbol prefix added per line by the caller. Returns s
// unchanged when c renders no foreground escape.
func PaintForeground(s string, c color.Color) string {
	prefix := foregroundCode(c)
	if prefix == "" {
		return s
	}

	s = sgrResetPattern.ReplaceAllStringFunc(s, func(reset string) string {
		return reset + prefix
	})

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
