package comment

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/bensadeh/circumflex/ansi"
	"github.com/bensadeh/circumflex/highlight"
	"github.com/bensadeh/circumflex/nerdfonts"
	"github.com/bensadeh/circumflex/style"
)

// RenderOptions carries the geometry- and presentation-dependent inputs of a
// render. Everything content-dependent lives in the blocks.
type RenderOptions struct {
	// CommentWidth is the prose column.
	CommentWidth int
	// ScreenWidth caps how far a code box may grow past CommentWidth.
	ScreenWidth int
	NerdFonts   bool
	// Fg tints the paragraphs (mod comments); nil renders untinted.
	Fg color.Color
}

const maxURLDisplay = 50

// baseStyle is the SGR context a block establishes around its spans: faint
// italic for quotes, the mod tint for moderator paragraphs. Every styled
// span re-opens it after its own closing reset, so the context survives
// embedded links and code; lipgloss.Wrap re-opens it on wrapped lines.
type baseStyle struct {
	open string
}

// RenderBlocks renders a parsed comment body. Blocks that render empty are
// skipped rather than leaving blank gaps.
func RenderBlocks(blocks []Block, opts RenderOptions) string {
	var parts []string

	for i := range blocks {
		if rendered := renderBlock(&blocks[i], opts); rendered != "" {
			parts = append(parts, rendered)
		}
	}

	return strings.Join(parts, "\n\n")
}

func renderBlock(b *Block, opts RenderOptions) string {
	switch b.kind {
	case blockRemoved:
		return style.Faint(b.text)

	case blockQuote:
		return renderQuote(b.spans, opts)

	case blockCode:
		return renderCode(b, opts.CommentWidth, opts.ScreenWidth)

	case blockParagraph:
		return renderParagraph(b.spans, opts)

	default:
		return ""
	}
}

func renderParagraph(spans []span, opts RenderOptions) string {
	if len(spans) == 0 {
		return ""
	}

	base := baseStyle{}
	if opts.Fg != nil {
		base.open = style.ForegroundCode(opts.Fg)
	}

	var sb strings.Builder

	sb.WriteString(base.open)

	for i := range spans {
		sb.WriteString(renderSpan(&spans[i], base, opts.NerdFonts))
	}

	if base.open != "" {
		sb.WriteString(ansi.Reset)
	}

	return lipgloss.Wrap(sb.String(), opts.CommentWidth, "")
}

// Quotes render faint italic behind a ▎ gutter. Embedded links re-open the
// quote style after themselves, so the text around them stays quiet.
func renderQuote(spans []span, opts RenderOptions) string {
	if len(spans) == 0 {
		return ""
	}

	base := baseStyle{open: ansi.Italic + ansi.Faint}

	var sb strings.Builder

	sb.WriteString(base.open)

	for i := range spans {
		sb.WriteString(renderSpan(&spans[i], base, opts.NerdFonts))
	}

	sb.WriteString(ansi.Reset)

	padStr := ansi.Faint + " " + style.IndentSymbol
	padWidth := lipgloss.Width(padStr)
	wrapped := lipgloss.Wrap(sb.String(), opts.CommentWidth-padWidth, "")

	return style.PrefixLines(wrapped, padStr)
}

// dedent drops the whitespace indent shared by every line. HN marks code by
// leading spaces, so the block always arrives indented; the box already sets
// it apart, and keeping the indent would pad the box's inside.
func dedent(text string) string {
	lines := strings.Split(text, "\n")

	indent := -1

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if trimmed == "" {
			continue
		}

		if lineIndent := len(line) - len(trimmed); indent < 0 || lineIndent < indent {
			indent = lineIndent
		}
	}

	if indent <= 0 {
		return text
	}

	for i, line := range lines {
		if len(line) >= indent {
			lines[i] = line[indent:]
		}
	}

	return strings.Join(lines, "\n")
}

// guessLang names a code block's language from its dedented text — the same
// form renderCode displays, so shebang detection sees column zero.
func guessLang(text string) string {
	return highlight.GuessLang(dedent(strings.Trim(text, "\n")))
}

// The box spans at least commentWidth and grows with long code lines up to
// screenWidth (the space left of the scrollbar).
func renderCode(b *Block, commentWidth, screenWidth int) string {
	// Tokenizing is width-independent and costs real time on big blocks, so
	// it runs once per block, not once per resize step.
	if !b.hlDone {
		b.hlOut = highlight.Code(dedent(strings.Trim(b.text, "\n")), b.lang)
		b.hlDone = true
	}

	if b.hlOut != "" {
		return highlight.Boxed(b.hlOut, b.lang, screenWidth, commentWidth)
	}

	content := dedent(strings.Trim(b.text, "\n"))
	wrapped := style.WrapWithin(content, screenWidth-style.RoundedBoxChrome)
	lines := strings.Split(wrapped, "\n")

	for i, line := range lines {
		lines[i] = ansi.Faint + line + ansi.Reset
	}

	return style.RoundedBox(strings.Join(lines, "\n"), commentWidth, "")
}

// renderSpan is the single place a span role becomes bytes. Every styled
// role ends in a reset of some form, so each re-opens the enclosing base
// style afterwards.
func renderSpan(s *span, base baseStyle, nerdFonts bool) string {
	switch s.format {
	case spanPlain:
		return s.text

	case spanItalic:
		return ansi.Italic + s.text + ansi.Reset + base.open

	case spanLink:
		return style.CommentURL(truncateURL(stripScheme(s.text)), s.href) + base.open

	case spanCodeInline:
		return style.CommentBacktick(s.text) + base.open

	case spanMention:
		if IsMod(strings.TrimPrefix(strings.TrimPrefix(s.text, " "), "@")) {
			return style.CommentMod(s.text) + base.open
		}

		return style.CommentMention(s.text) + base.open

	case spanVariable:
		return style.CommentVariable(s.text) + base.open

	case spanReference:
		return "[" + referenceDigits(s.text) + "]" + base.open

	case spanAbbreviation:
		if s.text == "IANAL" {
			return style.Red(s.text) + base.open
		}

		return style.Green(s.text) + base.open

	case spanYCLabel:
		return ycLabel(s.text, nerdFonts) + base.open

	default:
		return s.text
	}
}

func stripScheme(url string) string {
	if _, after, found := strings.Cut(url, "://"); found {
		return after
	}

	return url
}

func truncateURL(display string) string {
	runes := []rune(display)
	if len(runes) <= maxURLDisplay {
		return display
	}

	return string(runes[:maxURLDisplay]) + "…"
}

// RenderContent renders a comment body behind its depth-colored indent
// symbol. The symbol's column is carved out of ScreenWidth before rendering —
// code boxes grow up to the full width they are given, so it must already
// exclude that column or the wrap below breaks border lines.
func RenderContent(blocks []Block, depth int, opts RenderOptions) string {
	coloredIndentSymbol := colorizeIndentSymbol(style.IndentSymbol, depth)
	padWidth := lipgloss.Width(coloredIndentSymbol)

	opts.ScreenWidth -= padWidth

	formattedComment := RenderBlocks(blocks, opts)

	wrapped := lipgloss.Wrap(formattedComment, opts.ScreenWidth, "")

	return style.PrefixLines(wrapped, coloredIndentSymbol)
}

// colorizeIndentSymbol colors the ▎ gutter by nesting depth, cycling through
// the theme's indent palette. Top-level comments have no symbol.
func colorizeIndentSymbol(indentSymbol string, level int) string {
	if level == 0 {
		return ansi.Reset
	}

	cycle := style.IndentCycle()
	idx := (level - 1) % len(cycle)

	return ansi.Reset + cycle[idx](indentSymbol)
}

// referenceStyles is the single inventory of [N] footnote references: the
// tokenizer derives its literals from it and the renderer its colors, so the
// two cannot drift apart.
var referenceStyles = []struct {
	digits string
	color  func(string) string
}{
	{"0", style.White},
	{"1", style.Red},
	{"2", style.Yellow},
	{"3", style.Green},
	{"4", style.Blue},
	{"5", style.Cyan},
	{"6", style.Magenta},
	{"7", style.BrightWhite},
	{"8", style.BrightRed},
	{"9", style.BrightYellow},
	{"10", style.BrightGreen},
}

func referenceDigits(digits string) string {
	for _, r := range referenceStyles {
		if r.digits == digits {
			return r.color(digits)
		}
	}

	return digits
}

const ansiBlack = 16 // ANSI 256-color black

// ycLabel renders a "(YC W21)" occurrence: a colored season bar in nerd-fonts
// mode, colored text otherwise. The leading double reset in the nerd variant
// is historical — the replacement carried one and the bar builder another.
func ycLabel(label string, nerdFonts bool) string {
	c := style.HeadlineYCLabelColor()

	if !nerdFonts {
		return ansi.Reset + lipgloss.NewStyle().Foreground(c).Render(label)
	}

	const noBreakSpace = "\u00a0" // ties the glyph to its season

	season := strings.TrimPrefix(label, "YC ")
	text := nerdfonts.YCombinator + noBreakSpace + season

	content := lipgloss.NewStyle().
		Foreground(lipgloss.ANSIColor(ansiBlack)).
		Background(c).
		Render(text)

	border := lipgloss.NewStyle().Foreground(c)

	return ansi.Reset + ansi.Reset +
		border.Render(nerdfonts.LeftSeparator) +
		content +
		border.Render(nerdfonts.RightSeparator)
}
