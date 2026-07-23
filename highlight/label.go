package highlight

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/bensadeh/circumflex/style"
)

// brandColors approximates each language's identity color with the terminal
// palette's own colors — black and white excluded, they vanish against
// backgrounds. Languages without an entry label in faint.
var brandColors = map[string]lipgloss.Style{
	"Go":         lipgloss.NewStyle().Foreground(lipgloss.Blue),
	"Rust":       lipgloss.NewStyle().Foreground(lipgloss.Red),
	"Python":     lipgloss.NewStyle().Foreground(lipgloss.Yellow),
	"JavaScript": lipgloss.NewStyle().Foreground(lipgloss.BrightYellow),
	"TypeScript": lipgloss.NewStyle().Foreground(lipgloss.BrightBlue),
	"C":          lipgloss.NewStyle().Foreground(lipgloss.BrightBlue),
	"C++":        lipgloss.NewStyle().Foreground(lipgloss.BrightMagenta),
	"C#":         lipgloss.NewStyle().Foreground(lipgloss.Magenta),
	"Bash":       lipgloss.NewStyle().Foreground(lipgloss.Green),
	"Shell":      lipgloss.NewStyle().Foreground(lipgloss.Green),
	"Ruby":       lipgloss.NewStyle().Foreground(lipgloss.BrightRed),
	"PHP":        lipgloss.NewStyle().Foreground(lipgloss.Magenta),
	"Perl":       lipgloss.NewStyle().Foreground(lipgloss.Cyan),
	"Java":       lipgloss.NewStyle().Foreground(lipgloss.Red),
	"Kotlin":     lipgloss.NewStyle().Foreground(lipgloss.Magenta),
	"Swift":      lipgloss.NewStyle().Foreground(lipgloss.BrightRed),
	"Haskell":    lipgloss.NewStyle().Foreground(lipgloss.Magenta),
	"Zig":        lipgloss.NewStyle().Foreground(lipgloss.BrightYellow),
	"HTML":       lipgloss.NewStyle().Foreground(lipgloss.BrightRed),
	"CSS":        lipgloss.NewStyle().Foreground(lipgloss.BrightBlue),
	"Docker":     lipgloss.NewStyle().Foreground(lipgloss.BrightCyan),
	"JSX":        lipgloss.NewStyle().Foreground(lipgloss.BrightCyan),
	"TSX":        lipgloss.NewStyle().Foreground(lipgloss.BrightBlue),
}

// Label renders a detected language's display name for the code box header:
// the lexer's proper name, brand-colored when it has an identity color, faint
// otherwise, empty when lang names no lexer.
func Label(lang string) string {
	lexer := resolve(lang)
	if lexer == nil {
		return ""
	}

	name := lexer.Config().Name

	switch {
	case name == "Bash Session":
		name = "Shell"

	case name == "react":
		name = "JSX"

	case name == "EmacsLisp":
		name = "Emacs Lisp"

	// The generic alias claims the family and nothing more, so a guess with no
	// dialect evidence borrows the Common Lisp lexer without borrowing its name.
	case name == "Common Lisp" && strings.EqualFold(lang, "lisp"):
		name = "Lisp"

	case name == "markdown":
		name = "Markdown"

	case strings.EqualFold(lang, "tsx"):
		name = "TSX"
	}

	if s, ok := brandColors[name]; ok {
		return s.Render(name)
	}

	return style.Faint(name)
}
