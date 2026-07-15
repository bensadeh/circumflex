package style

import (
	"image/color"
	"strings"

	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/theme"

	"charm.land/lipgloss/v2"
	xansi "github.com/charmbracelet/x/ansi"
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

// RoundedBoxChrome is the width RoundedBox adds around its widest content
// line: two border cells and two padding cells. Callers wrap content at
// their available width minus this before boxing.
const RoundedBoxChrome = 4

// RoundedBox frames pre-wrapped content in a faint rounded border with one
// cell of horizontal padding. The frame spans at least width cells and grows
// with the content's widest line.
func RoundedBox(content string, width int) string {
	lines := strings.Split(content, "\n")

	inner := max(0, width-RoundedBoxChrome)
	for _, line := range lines {
		inner = max(inner, lipgloss.Width(line))
	}

	var b strings.Builder

	b.WriteString(Faint("╭" + strings.Repeat("─", inner+2) + "╮"))

	for _, line := range lines {
		pad := strings.Repeat(" ", inner-lipgloss.Width(line))
		b.WriteString("\n" + Faint("│") + " " + line + pad + " " + Faint("│"))
	}

	b.WriteString("\n" + Faint("╰"+strings.Repeat("─", inner+2)+"╯"))

	return b.String()
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
	commentURLStyle          lipgloss.Style // hyperlink applied at call time
	commentMentionStyle      lipgloss.Style
	commentModStyle          lipgloss.Style
	commentVariableStyle     lipgloss.Style
	commentOPStyle           lipgloss.Style
	commentGPStyle           lipgloss.Style
	commentNewIndicatorStyle lipgloss.Style
	commentBacktickStyle     lipgloss.Style

	metaAuthorStyle           lipgloss.Style
	metaScoreStyle            lipgloss.Style
	metaNewCommentsFaintStyle lipgloss.Style
	metaURLStyle              lipgloss.Style // hyperlink applied at call time

	readerH1Style lipgloss.Style
	readerH2Style lipgloss.Style
	readerH3Style lipgloss.Style
	readerH4Style lipgloss.Style
	readerH5Style lipgloss.Style
	readerH6Style lipgloss.Style

	readerLinkOpen string // SGR sequence opening a reader link: underline plus the theme color

	logoCStyle lipgloss.Style
	logoLStyle lipgloss.Style
	logoXStyle lipgloss.Style

	logoCFaintStyle lipgloss.Style
	logoLFaintStyle lipgloss.Style
	logoXFaintStyle lipgloss.Style
)

// Theme-dependent colors — rebuilt by rebuildThemeStyles whenever the theme changes.
var (
	headlineYCLabelColor  color.Color
	headlineYearColor     color.Color
	headlineAskHNColor    color.Color
	headlineShowHNColor   color.Color
	headlineTellHNColor   color.Color
	headlineThankHNColor  color.Color
	headlineLaunchHNColor color.Color
	headlineAudioColor    color.Color
	headlineVideoColor    color.Color
	headlinePDFColor      color.Color

	commentModColor color.Color

	readerImageColor color.Color

	headerCColor         color.Color
	headerLColor         color.Color
	headerXColor         color.Color
	headerPrimaryColor   color.Color
	headerSecondaryColor color.Color
	headerTertiaryColor  color.Color
)

var (
	indentCycleFuncs      []func(string) string
	indentCycleFaintFuncs []func(string) string
)

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

	headlineYCLabelColor = theme.ParseColor(current.Headline.YCLabel)
	headlineYearColor = theme.ParseColor(current.Headline.Year)
	headlineAskHNColor = theme.ParseColor(current.Headline.AskHN)
	headlineShowHNColor = theme.ParseColor(current.Headline.ShowHN)
	headlineTellHNColor = theme.ParseColor(current.Headline.TellHN)
	headlineThankHNColor = theme.ParseColor(current.Headline.ThankHN)
	headlineLaunchHNColor = theme.ParseColor(current.Headline.LaunchHN)
	headlineAudioColor = theme.ParseColor(current.Headline.Audio)
	headlineVideoColor = theme.ParseColor(current.Headline.Video)
	headlinePDFColor = theme.ParseColor(current.Headline.PDF)

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
	metaScoreStyle = fg(current.Meta.Score)
	metaNewCommentsFaintStyle = fg(current.Meta.NewComments).Faint(true)
	metaURLStyle = fg(current.Meta.URL)

	readerH1Style = fg(current.Reader.H1)
	readerH2Style = fg(current.Reader.H2)
	readerH3Style = fg(current.Reader.H3)
	readerH4Style = fg(current.Reader.H4)
	readerH5Style = fg(current.Reader.H5)
	readerH6Style = fg(current.Reader.H6)
	readerImageColor = theme.ParseColor(current.Reader.Image)
	readerLinkOpen = linkOpenSequence(theme.ParseColor(current.Reader.Link))

	searchMatchSGR = matchOverlaySGR(theme.ParseColor(current.Search.Match))
	searchCurrentSGR = currentOverlaySGR(theme.ParseColor(current.Search.Current))

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
	indentCycleFaintFuncs = make([]func(string) string, len(current.Indent.Cycle))

	for i, c := range current.Indent.Cycle {
		s := fg(c)
		faint := s.Faint(true)
		indentCycleFuncs[i] = func(text string) string { return s.Render(text) }
		indentCycleFaintFuncs[i] = func(text string) string { return faint.Render(text) }
	}
}

// A NoColor theme value means the terminal's default foreground, which is
// already in effect, so only the underline is emitted.
func linkOpenSequence(c color.Color) string {
	s := xansi.Style{}.Underline(true)

	if _, noColor := c.(lipgloss.NoColor); !noColor {
		s = s.ForegroundColor(c)
	}

	return s.String()
}

func CommentModFg() color.Color { return commentModColor }

// Dimmed is SGR faint done in color: fg blended halfway toward bg. For text
// that must read dim without the faint flag — terminals (Ghostty among them)
// dim underline and strikethrough decorations along with the glyphs, so a
// run that pins its underline color has to dim its text through the
// foreground instead.
func Dimmed(fg, bg color.Color) color.Color {
	mid := func(a, b uint32) uint8 { return uint8((a + b) / 2 >> 8) }

	fr, fgn, fb, _ := fg.RGBA()
	br, bgn, bb, _ := bg.RGBA()

	return color.RGBA{R: mid(fr, br), G: mid(fgn, bgn), B: mid(fb, bb), A: 255}
}

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

func HeadlineYCLabelColor() color.Color  { return headlineYCLabelColor }
func HeadlineYearColor() color.Color     { return headlineYearColor }
func HeadlineAskHNColor() color.Color    { return headlineAskHNColor }
func HeadlineShowHNColor() color.Color   { return headlineShowHNColor }
func HeadlineTellHNColor() color.Color   { return headlineTellHNColor }
func HeadlineThankHNColor() color.Color  { return headlineThankHNColor }
func HeadlineLaunchHNColor() color.Color { return headlineLaunchHNColor }
func HeadlineAudioColor() color.Color    { return headlineAudioColor }
func HeadlineVideoColor() color.Color    { return headlineVideoColor }
func HeadlinePDFColor() color.Color      { return headlinePDFColor }

func CommentURL(s, url string) string     { return commentURLStyle.Hyperlink(url).Render(s) }
func CommentMention(s string) string      { return commentMentionStyle.Render(s) }
func CommentMod(s string) string          { return commentModStyle.Render(s) }
func CommentVariable(s string) string     { return commentVariableStyle.Render(s) }
func CommentOP(s string) string           { return commentOPStyle.Render(s) }
func CommentGP(s string) string           { return commentGPStyle.Render(s) }
func CommentNewIndicator(s string) string { return commentNewIndicatorStyle.Render(s) }
func CommentBacktick(s string) string     { return commentBacktickStyle.Render(s) }

// CommentBacktickLink is the backtick style for inline code that is itself a
// hyperlink: the code keeps its color and the underline marks the link — the
// link wrapper's own underline is lost to the reset the backtick style needs.
func CommentBacktickLink(s string) string {
	return commentBacktickStyle.Underline(true).Render(s)
}

func MetaAuthor(s string) string           { return metaAuthorStyle.Render(s) }
func MetaScore(s string) string            { return metaScoreStyle.Render(s) }
func MetaNewCommentsFaint(s string) string { return metaNewCommentsFaintStyle.Render(s) }
func MetaURL(s, url string) string         { return metaURLStyle.Hyperlink(url).Render(s) }

func ReaderH1(s string) string      { return readerH1Style.Render(s) }
func ReaderH2(s string) string      { return readerH2Style.Render(s) }
func ReaderH3(s string) string      { return readerH3Style.Render(s) }
func ReaderH4(s string) string      { return readerH4Style.Render(s) }
func ReaderH5(s string) string      { return readerH5Style.Render(s) }
func ReaderH6(s string) string      { return readerH6Style.Render(s) }
func ReaderImageColor() color.Color { return readerImageColor }

// ReaderLink renders s as a clickable link: underlined, in the theme's link
// color, wrapped in an OSC 8 hyperlink to url. It closes with targeted
// resets (default foreground, underline off) instead of a full reset, so a
// link inside an otherwise styled run leaves the surrounding style intact.
func ReaderLink(s, url string) string {
	return ansi.Hyperlink(url, readerLinkOpen+s+ansi.DefaultForeground+ansi.UnderlineOff)
}

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

func IndentCycleFaint() []func(string) string { return indentCycleFaintFuncs }

// ForegroundCode returns the raw ANSI foreground escape for a color.Color,
// with no trailing reset. Returns "" when the color renders no escape.
func ForegroundCode(c color.Color) string {
	if c == nil {
		return ""
	}

	if _, noColor := c.(lipgloss.NoColor); noColor {
		return ""
	}

	return xansi.Style{}.ForegroundColor(c).String()
}
